package chisel

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func WriteResources(args Args) error {
	resourceDir := filepath.Join(args.ProjectDirectory, "libs", args.Language)
	err := filepath.Walk(resourceDir, func(path string, info os.FileInfo, err error) error {
		prefixRemoved := strings.TrimPrefix(path, resourceDir)
		if prefixRemoved == "" {
			return nil
		}
		fullPath := filepath.Join(args.Directory, prefixRemoved)
		dir := filepath.Dir(fullPath)
		e := os.MkdirAll(dir, 0755)
		if e != nil {
			return e
		}

		file, err := openFile(fullPath, path, args.Overwrite)
		if err != nil {
			return err
		}
		if file == nil {
			return nil
		}
		defer file.Close()

		templateFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer templateFile.Close()

		_, err = io.Copy(file, templateFile)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func openFile(path string, orig string, overwrite bool) (*os.File, error) {
	info, e := os.Stat(orig)
	if e != nil {
		return nil, e
	}
	if info.IsDir() {
		err := os.Mkdir(path, 0777)
		if err != nil && os.IsExist(err) {
			return nil, nil
		}
		return nil, err
	}

	_, err := os.Stat(path)
	if !overwrite && (!errors.Is(err, os.ErrNotExist) || errors.Is(err, os.ErrPermission)) {
		var response string
		fmt.Printf("%s already exists. Overwrite? [y/n] ", path)
		fmt.Scan(&response)
		response = strings.Trim(response, " \t\n")
		if response != "y" && response != "Y" {
			return nil, nil
		}
	}
	return os.Create(path)
}
