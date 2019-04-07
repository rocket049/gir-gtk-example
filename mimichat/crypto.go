package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
	"log"
)

var B64 *base64.Encoding

type MsgCrypto struct {
	key      []byte
	aesIV    []byte
	aesBlock cipher.Block
}

func init() {
	//log.Println("Init crypto")
	B64 = base64.NewEncoding("*$abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890").WithPadding(rune('%'))
}

//SetKey : Must use this method to set aes key
func (p *MsgCrypto) SetKey(k string) error {
	key := []byte(k)
	if len(key) == 32 {
		p.key = key
	} else if len(key) > 32 {
		p.key = key[0:32]
	} else {
		buf := bytes.NewBufferString("")
		n := 32 / len(key)
		m := 32 % len(key)
		for i := 0; i < n; i++ {
			buf.Write(key)
		}
		buf.Write(key[:m])
		p.key = buf.Bytes()
	}
	return p.initAESCipher()
}

//initAESCipher :initial block and encoder
func (p *MsgCrypto) initAESCipher() error {
	var err error
	p.aesBlock, err = aes.NewCipher(p.key)
	if err != nil {
		log.Println("makeEncoder:", err)
		return err
	}
	p.aesIV = make([]byte, p.aesBlock.BlockSize())
	return nil
}

//aes256Encode use private key
func (p *MsgCrypto) aes256Encode(msg []byte) []byte {
	if p.key == nil {
		log.Println("AES256Encode: Must use SetKey to set aes key")
		return nil
	}
	if len(msg) == 0 {
		return nil
	}
	io.ReadFull(rand.Reader, p.aesIV)
	aesEncoder := cipher.NewCTR(p.aesBlock, p.aesIV)
	var res = make([]byte, len(msg))
	aesEncoder.XORKeyStream(res, msg)
	buf1 := bytes.NewBufferString("")
	buf1.Write(p.aesIV)
	buf1.Write(res)
	return buf1.Bytes()
}

//aes256Decode use private key
func (p *MsgCrypto) aes256Decode(msg []byte) []byte {
	if p.key == nil {
		log.Println("AES256Decode: Must use SetKey to set aes key")
		return nil
	}
	bs := p.aesBlock.BlockSize()
	if len(msg) <= bs {
		log.Println("AES256Decode: message to short")
		return nil
	}
	aseDecoder := cipher.NewCTR(p.aesBlock, msg[:bs])
	var res = make([]byte, len(msg)-bs)
	aseDecoder.XORKeyStream(res, msg[bs:])
	return res
}

//Encode :加密 -> base64编码
func (p *MsgCrypto) Encode(msg []byte) string {
	aesArray := p.aes256Encode(msg)
	if aesArray == nil {
		return ""
	}
	res := B64.EncodeToString(aesArray)
	return res
}

//Decode :base64解码 -> 解密
func (p *MsgCrypto) Decode(msg string) []byte {
	aesArray, err := B64.DecodeString(msg)
	if err != nil {
		//log.Printf("Base64 Decode : %v\n", err)
		return nil
	}
	rawArray := p.aes256Decode(aesArray)
	return rawArray
}
