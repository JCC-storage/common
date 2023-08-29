package magefiles

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/sh"
	cp "github.com/otiai10/copy"
)

var Global = struct {
	OS        string
	Arch      string
	BuildRoot string
}{
	Arch: "amd64",
}

type BuildArgs struct {
	OutputName string
	OutputDir  string
	AssetsDir  string
	EntryFile    string
}

type goBuildArgs struct {
	Env       map[string]string
	OutputExt string
}

func Build(args BuildArgs) error {
	buildRoot := Global.BuildRoot
	if buildRoot == "" {
		buildRoot = "build"
	}

	fullOutputDir, err := filepath.Abs(filepath.Join(buildRoot, args.OutputDir))
	if err != nil {
		return err
	}

	goBuildArgs, err := makeGoBuildArgs()
	if err != nil {
		return err
	}

	binPath := filepath.Join(fullOutputDir, args.OutputName+goBuildArgs.OutputExt)
	fmt.Printf("building to %s\n", binPath)

	goCmdArgs := []string{ "build", "-o", binPath}
	if args.EntryFile != "" {
		goCmdArgs = append(goCmdArgs, args.EntryFile)
	}

	err = sh.RunWith(goBuildArgs.Env,"go", goCmdArgs...)
	if err != nil {
		return err
	}

	if args.AssetsDir != "" {
		outputAssetsPath := fullOutputDir
		fmt.Printf("copying asset to %s\n", outputAssetsPath)

		return CopyAssets(args.AssetsDir, outputAssetsPath)
	}

	return nil
}

func makeGoBuildArgs() (goBuildArgs, error) {
	args := goBuildArgs{
		Env: make(map[string]string),
	}

	if Global.OS == "win" {
		args.OutputExt = ".exe"
		args.Env["CGO_ENABLE"] = "0"
		args.Env["GOOS"] = "windows"
	} else if Global.OS == "linux" {
		args.OutputExt = ""
		args.Env["CGO_ENABLE"] = "0"
		args.Env["GOOS"] = "linux"
	} else if Global.OS != "" {
		return goBuildArgs{}, fmt.Errorf("unknow os type: %s", Global.OS)
	}

	var pltParts []string
	if Global.OS != "" {
		pltParts = append(pltParts, Global.OS)
	}
	if Global.Arch != "" {
		pltParts = append(pltParts, Global.Arch)
	}

	if len(pltParts) == 0 {
		fmt.Print("building platform is not set, will build for current machine.\n")
	} else {
		fmt.Printf("building for %s.\n", strings.Join(pltParts, "-"))
	}

	return args, nil
}

func CopyAssets(assrtDir string, targetDir string) error {
	info, err := os.Stat(assrtDir)
	if errors.Is(err, os.ErrNotExist) || !info.IsDir() {
		return nil
	}

	return cp.Copy(assrtDir, targetDir)
}
