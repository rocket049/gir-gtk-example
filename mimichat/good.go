//+build good

package main

import "C"

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/linuxdeepin/go-gir/gobject-2.0"

	"github.com/linuxdeepin/go-gir/gtk-3.0"
	"github.com/rocket049/gettext-go/gettext"
)

func init() {
	rand.Seed(time.Now().UnixNano())
	verbose = true
	step = 1
	exe1, _ := os.Executable()
	dir1 := filepath.Dir(exe1)
	gettext.BindTextdomain("mimichat", filepath.Join(dir1, "locale"), nil)
	gettext.Textdomain("mimichat")
}

//T translate
func T(s string) string {
	return gettext.PGettext("", s)
}

var (
	buf     string
	bot     *Bot
	step    int
	verbose bool
)

type chatWindow struct {
	win         gtk.Window
	grid        gtk.Grid
	msgView     gtk.TreeView
	msgStore    gtk.ListStore
	userView    gtk.TreeView
	userStore   gtk.ListStore
	inputBox    gtk.Entry
	userMsg     gtk.Label
	target      gtk.Label
	scrollLeft  gtk.ScrolledWindow
	scrollRight gtk.ScrolledWindow
	msgRender   gtk.CellRenderer
}

//const 列表成员
const (
	MsgSender int = iota
	MsgText
	MsgColor
)

func (s *chatWindow) Create() {
	w := gtk.WindowNew(gtk.WindowTypeToplevel)
	s.win = gtk.WrapWindow(w.Ptr)
	s.win.SetTitle(T("Security IRC Chat"))
	s.win.SetIconName("stock_internet")
	w = gtk.GridNew()
	s.grid = gtk.WrapGrid(w.Ptr)
	s.win.Add(w)
	w = gtk.LabelNew(T("User Message"))
	s.userMsg = gtk.WrapLabel(w.Ptr)
	s.grid.Attach(w, 0, 0, 2, 1)
	w = gtk.ScrolledWindowNew(gtk.Adjustment{}, gtk.Adjustment{})
	s.scrollLeft = gtk.WrapScrolledWindow(w.Ptr)
	s.scrollLeft.SetSizeRequest(500, 400)
	s.scrollLeft.SetHexpand(true)
	s.scrollLeft.SetVexpand(true)
	s.scrollLeft.SetPolicy(gtk.PolicyTypeNever, gtk.PolicyTypeAutomatic)
	s.grid.Attach(w, 0, 1, 1, 1)
	w = gtk.TreeViewNew()
	s.msgView = gtk.WrapTreeView(w.Ptr)
	s.msgView.SetGridLines(gtk.TreeViewGridLinesBoth)
	s.scrollLeft.Add(s.msgView.Widget)
	w = gtk.ScrolledWindowNew(gtk.Adjustment{}, gtk.Adjustment{})
	s.scrollRight = gtk.WrapScrolledWindow(w.Ptr)
	s.scrollRight.SetSizeRequest(100, 400)
	s.scrollRight.SetVexpand(true)
	s.grid.Attach(s.scrollRight.Widget, 1, 1, 1, 1)
	w = gtk.TreeViewNew()
	s.userView = gtk.WrapTreeView(w.Ptr)
	s.scrollRight.Add(s.userView.Widget)
	w = gtk.EntryNew()
	s.inputBox = gtk.WrapEntry(w.Ptr)
	s.grid.Attach(s.inputBox.Widget, 0, 2, 1, 1)
	w = gtk.LabelNew("#")
	s.target = gtk.WrapLabel(w.Ptr)
	s.grid.Attach(s.target.Widget, 1, 2, 1, 1)
	s.grid.SetColumnSpacing(3)
	s.grid.SetRowSpacing(3)
	s.grid.ShowAll()

	s.setMsgView()
	s.setUserView()
	s.setSignals()

	s.win.Show()
}

func (s *chatWindow) valueInt(v int) gobject.Value {
	p := gobject.ValueNew()
	p.Init(gobject.TYPE_INT)
	p.SetInt(v)
	return p
}

func (s *chatWindow) setMsgView() {
	//initial s.msgView
	s.msgStore = gtk.ListStoreNew([]gobject.Type{gobject.TYPE_STRING, gobject.TYPE_STRING, gobject.TYPE_STRING})
	render := gtk.CellRendererTextNew()
	area := gtk.CellAreaBoxNew()
	area.Add(render)
	colSender := gtk.TreeViewColumnNewWithArea(area)
	colSender.SetTitle(T("Sender"))
	colSender.AddAttribute(render, "text", MsgSender)
	colSender.AddAttribute(render, "background", MsgColor)
	width, _ := s.scrollRight.GetSizeRequest()
	colSender.SetFixedWidth(width)
	render.SetProperty("wrap-width", s.valueInt(width))
	s.msgRender = gtk.CellRendererTextNew()
	wleft, _ := s.scrollLeft.GetSizeRequest()
	s.msgRender.SetProperty("wrap-width", s.valueInt(wleft-width-20))
	area = gtk.CellAreaBoxNew()
	area.Add(s.msgRender)
	colMsg := gtk.TreeViewColumnNewWithArea(area)
	colMsg.SetTitle(T("Message"))
	colMsg.AddAttribute(s.msgRender, "text", MsgText)
	colMsg.AddAttribute(s.msgRender, "background", MsgColor)

	s.msgView.SetTooltipText(T("Double click to copy message to input box."))
	s.msgView.SetGridLines(gtk.TreeViewGridLinesBoth)
	s.msgView.SetModel(s.msgStore.TreeModel())
	s.msgView.AppendColumn(colSender)
	s.msgView.AppendColumn(colMsg)
}

func (s *chatWindow) setUserView() {
	//initial s.userView
	s.userStore = gtk.ListStoreNew([]gobject.Type{gobject.TYPE_STRING, gobject.TYPE_STRING})
	render := gtk.CellRendererTextNew()
	width, _ := s.scrollRight.GetSizeRequest()
	render.SetProperty("wrap-width", s.valueInt(width))
	area := gtk.CellAreaBoxNew()
	area.Add(render)
	colUser := gtk.TreeViewColumnNewWithArea(area)
	colUser.SetTitle(T("Users"))
	colUser.AddAttribute(render, "text", 0)
	colUser.AddAttribute(render, "background", 1)

	s.userView.AppendColumn(colUser)
	s.userView.SetModel(s.userStore.TreeModel())
}

func (s *chatWindow) stringValue(v string) gobject.Value {
	p := gobject.ValueNew()
	p.Init(gobject.TYPE_STRING)
	p.SetString(v)
	return p
}

func (s *chatWindow) appendMsg(sender, msg, color string) {
	//var iter gtk.TreeIter
	gobject.IdleAdd(func() bool {
		p := s.msgStore.Append()
		s.msgStore.SetValue(p, MsgSender, s.stringValue(sender))
		s.msgStore.SetValue(p, MsgText, s.stringValue(msg))
		s.msgStore.SetValue(p, MsgColor, s.stringValue(color))

		gobject.IdleAdd(func() bool {
			vadj := s.scrollLeft.GetVadjustment()
			//size := vadj.GetValue() + vadj.GetPageIncrement()
			//fmt.Println("VAdjustment:", vadj.GetValue(), vadj.GetPageIncrement())
			vadj.SetValue(vadj.GetUpper())
			return false
		})

		return false
	})
}

func (s *chatWindow) appendUser(name, color string) {
	//var iter gtk.TreeIter
	gobject.IdleAdd(func() bool {
		p := s.userStore.Append()
		s.userStore.SetValue(p, 0, s.stringValue(name))
		s.userStore.SetValue(p, 1, s.stringValue(color))

		gobject.IdleAdd(func() bool {
			vadj := s.scrollRight.GetVadjustment()
			//size := vadj.GetValue() + vadj.GetPageIncrement()
			//fmt.Println("VAdjustment:", vadj.GetValue(), vadj.GetPageIncrement())
			vadj.SetValue(vadj.GetUpper())
			return false
		})
		return false
	})
}

func (s *chatWindow) setSignals() {
	s.win.Connect("destroy", func() {
		bot.Close()
		fmt.Println("main quit")
	})
	s.win.Connect("show", func() {
		gobject.IdleAdd(func() bool {
			s.getLoginInfo()
			return false
		})
	})
	s.inputBox.Connect("activate", func() {
		msg := s.inputBox.GetText()
		bot.PrivMsgTo(bot.Channel, msg)
		s.inputBox.SetText("")
		s.appendMsg(bot.Nick, msg, "#89D4E5")
	})
	s.msgView.Connect("row-activated", func() {
		ok, model, iter := s.msgView.GetSelection().GetSelected()
		//model := s.msgView.GetModel()
		//path1 := gtk.WrapTreePath(p)
		//ok, iter := model.GetIter(path1)
		if !ok {
			return
		}
		v0 := model.GetValue(iter, 0)
		v1 := model.GetValue(iter, 1)
		s0 := v0.GetString()
		s1 := v1.GetString()
		s.inputBox.SetText(fmt.Sprintf("%s:%s", s0, s1))
	})
}

func (s *chatWindow) getLoginInfo() {
	w := gtk.DialogNew()
	dlg := gtk.WrapDialog(w.Ptr)
	dlg.SetTitle(T("Login"))
	dlg.SetTransientFor(s.win)
	w = gtk.GridNew()
	grid := gtk.WrapGrid(w.Ptr)
	w = gtk.LabelNew(T("Server:"))
	label := gtk.WrapLabel(w.Ptr)
	grid.Attach(label.Widget, 0, 0, 1, 1)
	w = gtk.EntryNew()
	server := gtk.WrapEntry(w.Ptr)
	server.SetText("chat.freenode.net:6667")
	server.SetWidthChars(25)
	grid.Attach(server.Widget, 1, 0, 1, 1)
	w = gtk.LabelNew(T("Room:"))
	label = gtk.WrapLabel(w.Ptr)
	grid.Attach(label.Widget, 0, 1, 1, 1)
	w = gtk.EntryNew()
	room := gtk.WrapEntry(w.Ptr)
	grid.Attach(room.Widget, 1, 1, 1, 1)
	w = gtk.LabelNew(T("Nick:"))
	label = gtk.WrapLabel(w.Ptr)
	grid.Attach(label.Widget, 0, 2, 1, 1)
	w = gtk.EntryNew()
	nick := gtk.WrapEntry(w.Ptr)
	nick.SetText(fmt.Sprintf("talker%d", rand.Uint32()))
	grid.Attach(nick.Widget, 1, 2, 1, 1)
	w = gtk.LabelNew(T("Key:"))
	label = gtk.WrapLabel(w.Ptr)
	grid.Attach(label.Widget, 0, 3, 1, 1)
	w = gtk.EntryNew()
	key := gtk.WrapEntry(w.Ptr)
	key.SetInputPurpose(gtk.InputPurposePassword)
	key.SetVisibility(false)
	grid.Attach(key.Widget, 1, 3, 1, 1)
	grid.ShowAll()

	dlg.AddButton(T("Login"), 1)
	dlg.AddButton(T("Cancel"), 2)
	dlg.SetChildVisible(true)
	child := dlg.GetChild()
	box := gtk.WrapBox(child.Ptr)
	box.PackStart(grid.Widget, true, true, 10)

	rid := dlg.Run()

	if rid == 1 {
		bot = new(Bot)
		bot.server = server.GetText()

		bot.User = nick.GetText()
		bot.Nick = bot.User

		room1 := room.GetText()

		bot.Channel = fmt.Sprintf("#xtalk%s", room1)
		s.target.SetText(bot.Channel)
		sk := key.GetText()

		bot.Crypto.SetKey(sk)
		step = 2
		s.userMsg.SetText(fmt.Sprintf("%s / %s / %s", bot.server, room1, bot.Nick))

		go s.doRecv()
	}
	dlg.Destroy()
	if step != 2 {
		gtk.MainQuit()
	}
}

//doRecv work in goroutine
func (s *chatWindow) doRecv() {
	bot.Connect()
	defer gtk.MainQuit()
	bot.Command("MODE", bot.Channel, "+s")
	for {
		message, err := bot.Recv()
		if err != nil {
			s.appendMsg("sys", "Exit Now!", "#FCBBBD")
			break
		}
		//fmt.Println(message.Command, message.Params)

		if message.Command == "JOIN" {
			//connected
			s.appendMsg(message.Prefix.Name, message.Command, "#21EFA9")
			if step < 3 && message.Prefix.Name == bot.Nick {
				step = 3
			}
			if message.Prefix.Name != bot.Nick {
				s.appendUser(message.Prefix.Name, "white")
			}
		} else if message.Command == "QUIT" {
			s.appendMsg(message.Prefix.Name, message.Command, "#FCBBBD")
			s.removeUser(message.Prefix.Name)
		} else if message.Command == "PING" {
			bot.Send(fmt.Sprintf("PONG %d", time.Now().UnixNano()))
			//log.Println("SEND: PONG")
		} else if message.Command == "PRIVMSG" {
			// Do Something with this msg
			rmsg := message.Params[1]
			dmsg := bot.Crypto.Decode(rmsg)
			name1 := message.Prefix.Name
			var msg string
			if dmsg == nil {
				msg = fmt.Sprintf("%s %s", T("[not secret]"), rmsg)
				s.appendMsg(string(name1), msg, "#FCBBBD")
			} else {
				msg = string(dmsg)
				s.appendMsg(string(name1), msg, "white")
			}
		} else if message.Command == "353" {
			names := strings.Split(message.Params[3], " ")
			for _, v := range names {
				s.appendUser(v, "white")
			}
		} else if verbose {
			//log.Printf("%v\n", message)
			msg := fmt.Sprintf("%s:%s", message.Command, strings.Join(message.Params, ","))
			s.appendMsg(message.Prefix.Name, msg, "#F9F9CB")
		}
	}
}

func (s *chatWindow) removeUser(name string) bool {
	model := s.userView.GetModel()
	ok, iter := model.GetIterFirst()
	if !ok {
		return false
	}
	v := model.GetValue(iter, 0)
	sv := v.GetString()
	if strings.Compare(sv, name) == 0 {
		return s.userStore.Remove(iter)
	}
	for model.IterNext(iter) {
		v := model.GetValue(iter, 0)
		sv := v.GetString()
		if strings.Compare(sv, name) == 0 {
			return s.userStore.Remove(iter)
		}
	}
	return false
}

func (s *chatWindow) Example() {
	s.appendMsg("tom", "hello", "white")
	s.appendMsg("jack", "hello", "#89D4E5")
	s.appendMsg("tom", "hello", "white")
	s.appendMsg("jack", "func (s *chatWindow) setSignals() func (s *chatWindow) setSignals() func (s *chatWindow) setSignals() ", "#89D4E5")
	s.appendMsg("jack", "func (s *chatWindow) setSignals() func (s *chatWindow) setSignals() func (s *chatWindow) setSignals() ", "white")
	s.appendMsg("jack", "func (s *chatWindow) setSignals() func (s *chatWindow) setSignals() func (s *chatWindow) setSignals() ", "#89D4E5")
	s.appendMsg("jack", "对于高性能分布式系统领域而言，Go 语言无疑比大多数其它语言有着更高的开发效率。它提供了海量并行的支持，这对于游戏服务端的开发而言是再好不过了。", "white")

	s.appendUser("jack", "#89D4E5")
	s.appendUser("tom", "white")
	s.appendUser("abcdefghijklmnopqrst", "#89D4E5")

	gobject.IdleAdd(func() bool {
		model := s.userView.GetModel()

		ok, iter := model.GetIterFirst()
		if !ok {
			return false
		}
		v := model.GetValue(iter, 0)
		sv := v.GetString()
		fmt.Println(sv)
		for model.IterNext(iter) {
			v := model.GetValue(iter, 0)
			sv := v.GetString()
			fmt.Println(sv)
		}
		return false
	})
}

func panicError(ok bool) {
	if !ok {
		panic(errors.New("panicError"))
	}
}

var (
	win1 *chatWindow
)

func main() {
	//defer gettext.SaveLog()
	gtk.Init0()
	win1 = new(chatWindow)
	win1.Create()
	gtk.Main()
}
