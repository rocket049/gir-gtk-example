## BUG报告
`gtk.TreeView.Connect("row-activated", CALLBACK )` 中的回调函数的参数无法使用。

### 这个程序用于测试 `go-gir`。

一个加密聊天软件，`AES256`加密，通过`IRC`服务器互联。

注：密码和聊天室随便填

linux编译，编译器go1.12：

1. `make good` -- 编译 `good` 程序，双击信息列表后，列表内容被复制到输入框内。
2. `make bad` -- 编译 `bad` 程序，双击信息列表无法复制内容到输入框。

注：两个程序的区别在 `good.go` 和 `bad.go` 的 229 行到 233 行。

```
//bad.go 第 229 行调用代码
	s.msgView.Connect("row-activated", func(v, p, c unsafe.Pointer) {
		//ok, model, iter := s.msgView.GetSelection().GetSelected()
		model := s.msgView.GetModel()
		path1 := gtk.WrapTreePath(p)
		ok, iter := model.GetIter(path1)
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

//good.go 第 229 行调用代码
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
```
