package main

import (
	"fmt"
	"io"
	"mime"
	"net/mail"
	"os"
	"strconv"
	"strings"

	"golang.org/x/net/html/charset"

	log "github.com/sirupsen/logrus"
)

const (
	singleAddrError    = "mail: expected single address"
	commaError         = "mail: expected comma"
	missingAtSignError = "mail: missing @ in addr-spec"
	invalidStrError    = "mail: invalid string"
	noAddrError        = "mail: no address"
	noAngleAddrError   = "mail: no angle-addr"
	missingSingsError  = "mail: missing '@' or angle-addr"
	unclosedAddrError  = "mail: unclosed angle-addr"
)

// createTmpDir создает временную директорию для хранения .eml файла письма.
// Возвращает имя созданной директории и ошибку, которая может возникнуть только при ее создании.
func createTmpDir() (string, error) {

	var err error
	name := os.TempDir() + "/" + stageName

	deleteFile(name)

	err = os.Mkdir(name, 0700)
	if os.IsExist(err) {
		log.Infof("Tmp dir %s was not deleted from the previous time.", name)
	}

	if os.IsNotExist(err) {
		if _, err = os.Stat(os.TempDir()); os.IsNotExist(err) {
			return "", fmt.Errorf("stat failed: %w", err)
		}
	}

	name += "/"
	return name, nil
}

// createFile создает .eml файл.
func createFile(id int64, tmpPath string) (*os.File, string, error) {

	path := tmpPath + strconv.FormatInt(id, 10) + ".eml"
	file, err := os.Create(path)
	if err != nil {
		return file, path, fmt.Errorf("create failed: %w", err)
	}

	return file, path, nil
}

// writeInFile записывает в файл данные письма.
func writeInFile(file *os.File, email Email, to, body string) error {
	defer file.Close()

	header := "From:" + strings.TrimSuffix(email.FromAddr, "\n") + "\n"
	header += "To:" + strings.ReplaceAll(to, "\n", "") + "\n"
	header += "Subject:" + strings.ReplaceAll(email.Subject, "\n", " ") + "\n"
	header += "Message-ID:" + strings.TrimSuffix(email.MessageID, "\n") + "\n"
	header += "Received:" + strings.TrimSuffix(email.Received, "\n") + "\n"
	header += "Date:" + strings.TrimSuffix(email.Date, "\n") + "\n"
	header += "MIME-Version:" + strings.TrimSuffix(email.MimeVersion, "\n") + "\n"
	header += "Content-Type:" + strings.TrimSuffix(email.ContentType, "\n") + "\n"
	header += "Content-Transfer-Encoding:" + strings.TrimSuffix(email.ContentTransferEncoding, "\n") + "\n\n"

	log.Debugf("writeInFile; header: %q", header)

	if _, err := file.WriteString(header + body); err != nil {
		return fmt.Errorf("write failed: %w", err)
	}

	return nil
}

// deleteFile удаляет файл по указанному пути.
func deleteFile(path string) {
	if err := os.RemoveAll(path); err != nil {
		log.Errorf("Failed to delete file (%s): %s", path, err)
	}
}

// cleanAddress очищает адрес от угловых скобок.
func cleanAddress(addr string) string {

	if len(addr) == 0 {
		return ""
	}

	splittedAddr := strings.Split(addr, "<")
	if len(splittedAddr) == 1 {
		return addr
	}

	return strings.TrimRight(splittedAddr[1], ">")
}

func charsetReader(label string, input io.Reader) (io.Reader, error) {

	label = strings.ReplaceAll(label, "windows-", "cp")
	encoding, _ := charset.Lookup(label)

	return encoding.NewDecoder().Reader(input), nil
}

// decodeHeader декодирует символы заголовка письма.
func decodeHeader(encodedHeader string) string {

	dec := mime.WordDecoder{CharsetReader: charsetReader}
	decodedHeader, err := dec.DecodeHeader(encodedHeader)
	if err != nil {
		return encodedHeader
	}

	return decodedHeader
}

// replaceRoundBracketsToAngle заменяет круглые скобки на угловые
// в адресе электронной почты.
func replaceRoundBracketsToAngle(addr string) string {

	addr = strings.ReplaceAll(addr, "(", "<")
	addr = strings.ReplaceAll(addr, ")", ">")
	return addr
}

func isExceptionError(err string) bool {
	return strings.Contains(err, singleAddrError) || strings.Contains(err, missingAtSignError) ||
		strings.Contains(err, noAngleAddrError) || strings.Contains(err, commaError) || strings.Contains(err, missingSingsError)
}

func trimAddrName(addr string) string {
	addr = strings.ReplaceAll(addr, " ", "")
	trimIndex := strings.Index(addr, "<")
	if trimIndex > 0 {
		return string([]rune(addr)[trimIndex-1:])
	}

	return addr
}

// parseAddresses парсит адреса из полей From и To заголовка письма.
func parseAddresses(from, to string) (string, string) {

	from = strings.TrimSpace(from)
	to = strings.TrimSpace(to)

	if strings.Contains(from, "\"") {
		from = strings.ReplaceAll(from, "\"", "")
	}

	if strings.Contains(from, "(") || strings.Contains(from, ")") {
		from = replaceRoundBracketsToAngle(from)
	}

	if strings.Contains(to, "(") || strings.Contains(to, ")") {
		to = replaceRoundBracketsToAngle(to)
	}

	if strings.Contains(to, ";") {
		to = strings.ReplaceAll(to, ";", ",")
	}

	parsedFrom, err := mail.ParseAddress(from)
	if err != nil && isExceptionError(err.Error()) {
		log.Errorf("Could not parse From address: %v", err)
		addrs := strings.Split(from, " ")

		parsedFrom, err = mail.ParseAddress(addrs[len(addrs)-1])
		if err != nil {
			log.Errorf("Could not parse From address: %v", err)
			return from, to
		}
	} else if err != nil && strings.Contains(err.Error(), unclosedAddrError) {
		from = trimAddrName(from)
		parsedFrom, err = mail.ParseAddress(from)
		if err != nil {
			log.Errorf("Could not parse From address: %v", err)
			return from, to
		}
	} else if err != nil {
		log.Errorf("Could not parse From address: %v", err)
		return from, to
	}
	from = parsedFrom.Address

	if strings.Contains(to, ",") {
		parsedTo, err := mail.ParseAddressList(to)
		if err != nil && (strings.Contains(err.Error(), commaError) || strings.Contains(err.Error(), missingAtSignError)) {
			log.Errorf("Could not parse To address list: %v", err)
			addrs := strings.Split(to, " ")
			var toCopy string
			addrsMap := make(map[string]bool, 0)
			for _, addr := range addrs {
				if (strings.Contains(addr, "<") || strings.Contains(addr, ">")) && strings.Contains(addr, "@") {
					addr = strings.TrimSuffix(addr, ",")
					if !addrsMap[addr] {
						toCopy += addr + ","
						addrsMap[addr] = true
					}
				}
			}

			toCopy = strings.TrimSuffix(toCopy, ",")
			parsedTo, err = mail.ParseAddressList(toCopy)
			if err != nil {
				log.Errorf("Could not parse To address list: %v", err)
				return from, to
			}
		} else if err != nil && strings.Contains(err.Error(), unclosedAddrError) {
			log.Errorf("Could not parse To address list: %v", err)
			to = strings.ReplaceAll(to, " >", ">")

			parsedTo, err = mail.ParseAddressList(to)
			if err != nil {
				log.Errorf("Could not parse To address list: %v", err)
				return from, to
			}
		} else if err != nil && strings.Contains(err.Error(), invalidStrError) {
			log.Errorf("Could not parse To address: %v", err)
			return from, ""
		} else if err != nil {
			log.Errorf("Could not parse To address list: %v", err)
			return from, to
		}

		addrs := ""
		for i, addr := range parsedTo {

			if i == len(parsedTo)-1 {
				addrs += addr.Address
				break
			}

			addrs += addr.Address + ","
		}

		to = addrs
	} else {
		parsedTo, err := mail.ParseAddress(to)
		if err != nil && (strings.Contains(err.Error(), singleAddrError) || strings.Contains(err.Error(), missingAtSignError)) {
			log.Errorf("Could not parse To address: %v", err)
			addrs := strings.Split(to, " ")

			parsedTo, err = mail.ParseAddress(addrs[len(addrs)-1])
			if err != nil {
				log.Errorf("Could not parse To address: %v", err)
				return from, to
			}
		} else if err != nil && strings.Contains(err.Error(), unclosedAddrError) {
			to = trimAddrName(to)
			parsedTo, err = mail.ParseAddress(to)
			if err != nil {
				log.Errorf("Could not parse To address: %v", err)
				return from, to
			}
		} else if err != nil && strings.Contains(err.Error(), noAddrError) {
			log.Errorf("Could not parse To address: %v", err)
			return from, ""
		} else if err != nil {
			log.Errorf("Could not parse To address: %v", err)
			return from, to
		}
		to = parsedTo.Address
	}

	return from, to
}
