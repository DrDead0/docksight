package release

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)


func ExtractTarGz(
	source string,
	destination string,
) error {


	file, err := os.Open(source)

	if err != nil {
		return err
	}

	defer file.Close()


	gzipReader, err := gzip.NewReader(file)

	if err != nil {
		return err
	}

	defer gzipReader.Close()


	tarReader := tar.NewReader(gzipReader)



	for {

		header, err := tarReader.Next()


		if err == io.EOF {
			break
		}


		if err != nil {
			return err
		}


		target := filepath.Join(
			destination,
			header.Name,
		)



		switch header.Typeflag {


		case tar.TypeDir:

			err := os.MkdirAll(
				target,
				0755,
			)

			if err != nil {
				return err
			}



		case tar.TypeReg:


			err := os.MkdirAll(
				filepath.Dir(target),
				0755,
			)

			if err != nil {
				return err
			}


			out, err := os.Create(target)

			if err != nil {
				return err
			}


			_, err = io.Copy(
				out,
				tarReader,
			)


			out.Close()


			if err != nil {
				return err
			}


		default:

			return fmt.Errorf(
				"unsupported archive entry: %s",
				header.Name,
			)
		}

	}


	return nil
}