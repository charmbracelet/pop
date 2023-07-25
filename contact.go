package main

import (
	"os"
	"strings"

	vcard "github.com/emersion/go-vcard"
)

func isContact(recipient *[]string) {
	for i, contact := range *recipient {
		if strings.HasSuffix(contact, ".vcf") || strings.HasSuffix(contact, ".vcard") {
			email, err := getEmail(contact)

			// Usually it's != nil, but if thrown an error we just leave this untouched
			// If logger added in the future, percolate error
			if err == nil && email != "" && recover() == nil {
				(*recipient)[i] = email
			}
		}
	}
}

func getEmail(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	unMarshaller := vcard.NewDecoder(file)

	card, err := unMarshaller.Decode()
	if err != nil {
		return "", err
	}
	email := card.PreferredValue(vcard.FieldEmail)

	return email, nil
}
