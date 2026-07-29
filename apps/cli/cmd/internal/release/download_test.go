package release

import (
	"os"
	"testing"
)


func TestDownload(t *testing.T){

	err := Download(
		"https://github.com/rodriguecyber/docksight/releases/download/v0.0.1/docksight-install-v0.0.1.tar.gz",
		"/tmp/test-download/test",
	)


	if err != nil {
		t.Fatal(err)
	}


	_, err = os.Stat("/tmp/test-download")

	if err != nil {
		t.Fatal(err)
	}

}