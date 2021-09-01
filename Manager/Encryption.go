package Manager

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Encrypt struct {
	inFile     string
	outFile    string
	key        string
	bufferSize int64
}

type Decrypt struct {
	inFile     string
	outFile    string
	key        string
	bufferSize int64
}

func (e *Encrypt) Encrypt(inFile string, outFile string, key string, bufferSize int64, checksumDir string) error {

	f, err := os.Open(inFile)

	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	k, err := ioutil.ReadFile(key)

	if err != nil {
		log.Fatal(err)
	}

	b, err := aes.NewCipher(k)

	if err != nil {
		log.Panic(err)
	}

	iv := make([]byte, b.BlockSize())
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		log.Fatal(err)
	}

	outfile, err := os.OpenFile(outFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer outfile.Close()

	// The buffer size must be multiple of 16 bytes

	log.Printf("Encrypting backup...")

	buf := make([]byte, bufferSize)
	stream := cipher.NewCTR(b, iv)
	for {
		n, err := f.Read(buf)
		if n > 0 {
			stream.XORKeyStream(buf, buf[:n])
			// Write into file
			_, err = outfile.Write(buf[:n])
			if err != nil {
				return err
			}
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Printf("Read %d bytes: %v", n, err)
			break
		}
	}
	_, err = outfile.Write(iv)
	if err != nil {
		return err
	}
	WriteChecksumToFile(checksumDir, CalculateChecksum(outFile))

	err = os.Remove(inFile)
	if err != nil {
		return err
	}

	return err
}

func (d *Decrypt) Decrypt(inFile string, outFile string, key string, bufferSize int64, checksumDir string, date string) error {

	if !ValidateChecksum(CalculateChecksum(inFile), checksumDir, filepath.Join(checksumDir, date)) {
		log.Fatal("Checksum validation failed! someone has tampered with the backup file!")
	} else {
		log.Printf("Checksum validation passed!")
	}

	f, err := os.Open(inFile)

	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	k, err := ioutil.ReadFile(key)

	if err != nil {
		log.Fatal(err)
	}

	b, err := aes.NewCipher(k)

	if err != nil {
		log.Panic(err)
	}

	fi, err := f.Stat()
	if err != nil {
		log.Fatal(err)
	}

	iv := make([]byte, b.BlockSize())
	msgLen := fi.Size() - int64(len(iv))
	_, err = f.ReadAt(iv, msgLen)
	if err != nil {
		log.Fatal(err)
	}

	outfile, err := os.OpenFile(outFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer outfile.Close()

	log.Printf("Decrypting backup...")

	// The buffer size must be multiple of 16 bytes
	buf := make([]byte, bufferSize)
	stream := cipher.NewCTR(b, iv)
	for {
		n, err := f.Read(buf)
		if n > 0 {
			// The last bytes are the IV, don't belong the original message
			if n > int(msgLen) {
				n = int(msgLen)
			}
			msgLen -= int64(n)
			stream.XORKeyStream(buf, buf[:n])
			// Write into file
			outfile.Write(buf[:n])
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Printf("Read %d bytes: %v", n, err)
			break
		}
	}
	err = os.Remove(inFile)
	if err != nil {
		return err
	}

	return err
}

func CalculateChecksum(file string) (checksum string) {

	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	log.Printf("Calculating checksum...")
	hash := md5.New()
	if _, err := io.Copy(hash, f); err != nil {
		return ""
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func WriteChecksumToFile(checksumDir string, checksum string) {

	currentTime := time.Now()
	// make sure that checksums directory exists
	if _, err := os.Stat(checksumDir); os.IsNotExist(err) {
		err := os.Mkdir(checksumDir, 0640)
		if err != nil {
			log.Printf("Unable to create checksum directory")
		}
	}

	// if file does not exist then create it
	if _, err := os.Stat(filepath.Join(checksumDir, currentTime.Format("2006-01-02"))); os.IsNotExist(err) {
		_, err := os.Create(filepath.Join(checksumDir, currentTime.Format("2006-01-02")))
		if err != nil {
			log.Printf("Unable to create checksum file")
		}
	} else {
		log.Printf("Checksum file already exists... overwriting...")
	}

	f, err := os.OpenFile(
		filepath.Join(checksumDir, currentTime.Format("2006-01-02")),
		os.O_APPEND|os.O_WRONLY|os.O_TRUNC, 0640)

	if err != nil {
		fmt.Println(err)
	}

	w, err := io.WriteString(f, checksum)

	if err != nil {
		fmt.Println(w, err)
	}

	err = f.Close()
	if err != nil {
		log.Printf("Unable to close checksum file")
	}

}

func ValidateChecksum(checksum string, checksumDir string, checksumFile string) (valid bool) {

	if _, err := os.Stat(checksumDir); os.IsNotExist(err) {
		log.Fatal("Checksum file does not exist")
		return false
	}

	if _, err := os.Stat(checksumFile); os.IsNotExist(err) {
		log.Fatal("Checksum file does not exist")
		return false
	}

	f, err := ioutil.ReadFile(checksumFile)

	if err != nil {
		log.Printf("Cannot read checksum file to validate checksum")
	}

	s := string(f)

	if strings.Contains(s, checksum) {
		return true
	} else {
		return false
	}

	return false
}
