package store_config

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"encoding/json"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

const (
	srcKey  = "src"
	destKey = "dest"
)

var log *zap.Logger

type Manifest struct {
	Name      string            `json:"name"`
	Providers map[string]string `json:"providers"`
	RuleSets  map[string]string `json:"rulesets"`
}

type RuleSet struct {
	ProviderName string `yaml:"loadExternalProvider"`
}

func InitCommand(zapLogger *zap.Logger) *cobra.Command {
	log = zapLogger
	CobraCommand := &cobra.Command{
		Use:   "build-store-configs",
		Short: "Lookup store configs then copy into CWS provider folder",
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.Wrap(proceed(), "proceed")
		},
	}

	CobraCommand.PersistentFlags().StringP(srcKey, "s", "", "Source directory which be located store configs (YAML & Provider)")
	CobraCommand.MarkFlagRequired(srcKey)
	viper.BindPFlag(srcKey, CobraCommand.PersistentFlags().Lookup(srcKey))

	CobraCommand.PersistentFlags().StringP(destKey, "d", "", "Destination directory which be stored store provider")
	CobraCommand.MarkFlagRequired(destKey)
	viper.BindPFlag(destKey, CobraCommand.PersistentFlags().Lookup(destKey))

	return CobraCommand
}

func proceed() error {
	srcDir := viper.GetString(srcKey)
	dstDir := viper.GetString(destKey)

	if stat, err := os.Stat(srcDir); os.IsNotExist(err) || !stat.IsDir() {
		return errors.New("Src directory is not exists")
	}

	if stat, err := os.Stat(dstDir); os.IsNotExist(err) || !stat.IsDir() {
		return errors.New("Dst directory is not exists")
	}

	files, err := filepath.Glob(path.Join(srcDir, "./*/manifest.json"))
	if err != nil {
		return errors.Wrap(err, "filepath.Glob")
	}

	if len(files) == 0 {
		log.Fatal("There are no store configs at source directory")
	}

	for _, file := range files {
		if err := proceedWithManifest(file); err != nil {
			return errors.Wrap(err, "proceedWithManifest")
		}
	}

	log.Debug("Completed to copy store providers")
	return nil
}

func proceedWithManifest(filePath string) error {
	zl := log.With(zap.String("filePath", filePath))
	zl.Debug("Get file, try to get provider information")

	dirPath := filepath.Dir(filePath)

	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		return errors.Wrap(err, "ioutil.ReadFile(filePath)")
	}

	var (
		manifest       Manifest
		ruleset        RuleSet
		rulesetDesktop string
	)
	json.Unmarshal(file, &manifest)

	if len(manifest.RuleSets) == 0 && len(manifest.Providers) == 0 {
		zl.Debug("Rulesets and Providers are empty, skip it")
		return nil
	}

	for k, v := range manifest.RuleSets {
		if strings.Contains(k, "desktop") {
			rulesetDesktop = v
		}
	}

	rulesetFile, err := ioutil.ReadFile(path.Join(dirPath, rulesetDesktop))
	if err != nil {
		return errors.Wrap(err, "ioutil.ReadFile(rulesetDesktop)")
	}
	yaml.Unmarshal(rulesetFile, &ruleset)
	fmt.Println(ruleset)
	if len(ruleset.ProviderName) == 0 {
		zl.Debug("There're no provider for desktop, skip it")
		return nil
	}

	providerPath, ok := manifest.Providers[ruleset.ProviderName]
	if !ok {
		zl.Debug("Provider is exposed but it's not defined")
		return nil
	}

	destPath := path.Join(viper.GetString(destKey), ruleset.ProviderName+".js")
	providerPath = path.Join(dirPath, providerPath)
	if errCopy := copyFileContents(providerPath, destPath); errCopy != nil {
		return errors.Wrap(errCopy, "copyFileContents")
	}

	return nil
}

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}
