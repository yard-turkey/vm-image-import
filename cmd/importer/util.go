package main

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/minio/minio-go"
	"io"
	"net/http"
	"os"
	"strings"
	"encoding/base64"
)

func parseEnvVar(envVarName string, decode bool) string {
	value := os.Getenv(envVarName)
	if decode {
		v, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			glog.Fatalf("Error decoding environment variable %q", envVarName)
		}
		value = fmt.Sprintf("%s", v)
	}
	return value
}

func getDataWithClient(con *importInfo) io.ReadCloser {
	mc, err := minio.NewV4(con.endpoint, con.accessKeyId, con.secretKey, false)
	if err != nil {
		glog.Fatalf("func getDataWithClient: Could not create Minio client: %v", err)
	}
	objPath := strings.Split(con.objectPath, "/")
	bucketName := objPath[0]
	objName := strings.Join(objPath[1:], "_")
	glog.Infof("Streaming object %s", con.objectPath)
	objectReader, err := mc.GetObject(bucketName, objName, minio.GetObjectOptions{})
	if err != nil {
		glog.Fatalf("func getDataWithClient: failed getting objectPath: %v", err)
	}
	return objectReader
}

func getDataWithHTTP(url string) io.ReadCloser {
	resp, err := http.Get(url)
	if err != nil {
		glog.Fatalf("func streamDataFromURL: response body error: %v", err)
	}
	return resp.Body
}