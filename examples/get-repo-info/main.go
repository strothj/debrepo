package main

import (
	"log"

	"github.com/strothj/debrepo"
	"github.com/strothj/hkp"

	"golang.org/x/net/context"
)

func main() {
	debianJessieArchiveSigningKey := "126C0D24BD8A2942CC7DF8AC7638D0442B90D010"
	ubuntuKeyServer := "keyserver.ubuntu.com"
	ctx := context.Background()

	ks, _ := hkp.ParseKeyserver(ubuntuKeyServer)
	keyID, _ := hkp.ParseKeyID(debianJessieArchiveSigningKey)
	hkpClient := hkp.NewClient(ks, nil)
	keyring, err := hkpClient.GetKeysByID(ctx, keyID)
	if err != nil {
		log.Printf("error getting key: %v\n", err)
	}
	if len(keyring) == 0 {
		log.Println("failed to get key")
	}

	source, _ := debrepo.ParseSource("deb http://ftp.debian.org/debian squeeze main contrib non-free")
	sourceList := debrepo.SourceList([]*debrepo.Source{source})
	_ = sourceList

	// client := debrepo.NewClient(sources, keyring, nil)
	// _ = client

}
