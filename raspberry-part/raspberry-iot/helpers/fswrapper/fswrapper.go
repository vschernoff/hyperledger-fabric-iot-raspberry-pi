package fswrapper

import (
	"hlf-iot/config"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type SendElementStructure struct {
	Data     string                     `json:"data"`
	Url      string                     `json:"url"`
	CheckFcn func(string) (bool, error) `json:"checkfcn"`
}

func GetFilesDataFromKeyStorage(keyExtension string) ([]config.Key, error) {
	keys, err := GetKeys(config.MEDIA_ROOT_PATH, keyExtension)
	if err != nil {
		return nil, err
	}

	return keys, nil
}

func GetKeys(path, extension string) ([]config.Key, error) {
	var keys []config.Key

	entities, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, entity := range entities {
		if entity.IsDir() {
			keysFromDir, err := GetKeys(path+entity.Name()+"/", extension)
			if err != nil {
				return nil, err
			}

			keys = append(keys, keysFromDir...)
		}

		files, err := filepath.Glob(path + "*." + extension)
		if err != nil {
			return nil, err
		}

		for _, file := range files {
			fileData, err := ReadFile(file)
			if err != nil {
				return nil, err
			}

			if len(fileData.PrivateKeyHash) > 0 && len(fileData.Certificate) > 0 {
				keys = append(keys, fileData)
			}
		}
	}

	return keys, nil
}

func ReadFile(path string) (config.Key, error) {
	var fileData config.Key

	file, err := os.Open(path)
	if err != nil {
		return fileData, err
	}
	defer file.Close()

	data := make([]byte, 64)
	fileString := ""

	for {
		n, err := file.Read(data)
		if err == io.EOF {
			break
		}

		fileString += string(data[:n])
	}

	_, fileName := filepath.Split(file.Name())
	fileData.PrivateKeyHash = config.FilenameWithoutExtension(fileName)
	fileData.Certificate = fileString

	return fileData, nil
}
