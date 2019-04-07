package main

import (
	//"encoding/hex"
	"fmt"
	"log"
	"net"

	"golang.org/x/net/proxy"
	"gopkg.in/sorcix/irc.v2"
)

type BotInterface interface {
	Connect(conn net.Conn, err error)
	Send(string)
}

type Bot struct {
	proxy         string
	server        string
	Nick          string
	User          string
	Channel       string
	pass          string
	pread, pwrite chan string
	rawconn       net.Conn
	conn          *irc.Conn
	Crypto        MsgCrypto
}

func (bot *Bot) Send(command string) {
	fmt.Fprintf(bot.rawconn, "%s \r\n", command)
}

func (bot *Bot) PrivMsgTo(target, msg string) error {
	msg1 := bot.Crypto.Encode([]byte(msg))
	if msg1 == "" {
		return nil
	}
	message := &irc.Message{
		Prefix: &irc.Prefix{
			Name: bot.Nick,
			User: bot.User,
			Host: bot.server,
		},
		Command: "PRIVMSG",
		Params:  []string{target, msg1},
	}
	//fmt.Println(message.String())
	_, err := bot.conn.Write(message.Bytes())
	return err
}
func (bot *Bot) Command(cmd, target, opts string) error {
	message := &irc.Message{
		Prefix: &irc.Prefix{
			Name: bot.Nick,
			User: bot.User,
			Host: bot.server,
		},
		Command: cmd,
		Params:  []string{target, opts},
	}
	//fmt.Println(message.String())
	_, err := bot.conn.Write(message.Bytes())
	return err
}

func NewBot(proxy string, server string, Nick string, User string, Channel string, pass string) *Bot {
	//var nick1 = "user" + hex.EncodeToString([]byte(Nick))
	nick1 := Nick
	log.Printf("Nick:%v / Channel:%v\n", nick1, Channel)
	return &Bot{
		proxy:   proxy,
		server:  server,
		Nick:    nick1,
		User:    User,
		Channel: Channel,
		pass:    pass,
		rawconn: nil,
		conn:    nil,
	}
}

//SetPass set the password
func (bot *Bot) SetPass(pass string) {
	bot.pass = pass
}

//Connect init connection bot.rawconn bot.conn
func (bot *Bot) Connect() (conn *irc.Conn, err error) {
	if len(bot.proxy) > 0 {
		dialer, err := proxy.SOCKS5("tcp", bot.proxy, nil, proxy.Direct)
		if err != nil {
			log.Fatal("unable to connect to proxy server ", err)
		}
		bot.rawconn, err = dialer.Dial("tcp", bot.server)
		if err != nil {
			log.Fatal("unable to connect to IRC server ", err)
		}
	} else {
		bot.rawconn, err = net.Dial("tcp", bot.server)
		if err != nil {
			log.Fatal("unable to connect to IRC server ", err)
		}
	}
	bot.conn = irc.NewConn(bot.rawconn)
	//log.Printf("Connected to IRC server %s (%s) \n", bot.server, bot.rawconn.RemoteAddr())
	if len(bot.pass) > 0 {
		bot.Send(fmt.Sprintf("PASS %s \r\n", bot.pass))
	}
	bot.Send(fmt.Sprintf("USER %s 8 * :%s", bot.User, bot.User))
	bot.Send(fmt.Sprintf("NICK %s \r\n", bot.Nick))
	bot.Send(fmt.Sprintf("JOIN %s \r\n", bot.Channel))
	if err != nil {
		return nil, err
	} else {
		return bot.conn, nil
	}
}

//Close close bot
func (bot *Bot) Close() {
	if bot.conn != nil {
		bot.conn.Close()
	}
}

//Recv receive message
func (bot *Bot) Recv() (*irc.Message, error) {
	return bot.conn.Decode()
}
