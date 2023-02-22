package tools

import (
	"fmt"
	"github.com/spf13/viper"
	"io"
	"log"
	"os"
)

type MyUint64List []uint64

func (my64 MyUint64List) Len() int           { return len(my64) }
func (my64 MyUint64List) Swap(i, j int)      { my64[i], my64[j] = my64[j], my64[i] }
func (my64 MyUint64List) Less(i, j int) bool { return my64[i] < my64[j] }

func CreateLogger(logFilePath, logFileName, level string, printScreen bool) *log.Logger {
	file := logFilePath + "/" + logFileName + ".txt"
	fmt.Printf("日志文件: %s\n", file)
	logFile, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
	if nil != err {
		fmt.Printf("%s file not exist", file)
		panic(err)
	}
	//writers := []io.Writer{logFile, os.Stdout}
	var writers []io.Writer
	if printScreen {
		writers = []io.Writer{logFile, os.Stdout}
	} else {
		writers = []io.Writer{logFile}
	}
	fileAndStdoutWrite := io.MultiWriter(writers...)
	logger := log.New(fileAndStdoutWrite, level, log.Ldate|log.Ltime|log.Lshortfile)
	logger.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	return logger
}

func LoadConfiguration(fileName string, filePath string, topNode string) *viper.Viper {
	viper.SetConfigType("yaml")
	viper.SetConfigName(fileName)
	viper.AddConfigPath(filePath)
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("load config file error: %s\n", err)
		os.Exit(1)
	}
	// 获取顶级节点下的信息
	infos := viper.Sub(topNode)
	return infos
}
