package main

import "testing"

func TestAftonbladetParseFull(t *testing.T) {
	const in = "<![CDATA[<img src=\"http://gfx.aftonbladet-cdn.se/image/16518937/250/widescreen/49db9ec601a95/Ger+sig+ut+p%C3%A5+ov%C3%A4ntat+djupt+vatten \" /><p>Se när han grundluras av polarna</p>]]>"
	const firstOut = "Se när han grundluras av polarna"
	const secondOut = "http://gfx.aftonbladet-cdn.se/image/16518937/250/widescreen/49db9ec601a95/Ger+sig+ut+p%C3%A5+ov%C3%A4ntat+djupt+vatten"
	if x, y := AftonbladetParse(in); x != firstOut || y != secondOut {
		t.Errorf("\nTestAftonbladetParse\n(%s)\n = \n(%s,\n%s)\n want \n(%s,\n%s)", in, x, y, firstOut, secondOut)
	}
}

func TestAftonbladetParseEmpty(t *testing.T) {
	const in = "<![CDATA[<p>Peter Kadhammars berättelse om ett Sverige som varit, och ett som blir. • För att läsa: Ladda ner dokumentet som pdf eller zooma in direkt på sidan.</p>]]>"
	const firstOut = "Peter Kadhammars berättelse om ett Sverige som varit, och ett som blir. • För att läsa: Ladda ner dokumentet som pdf eller zooma in direkt på sidan."
	const secondOut = ""
	if x, y := AftonbladetParse(in); x != firstOut || y != secondOut {
		t.Errorf("\nTestAftonbladetParse\n(%s)\n = \n(%s,\n%s)\n want \n(%s,\n%s)\n", in, x, y, firstOut, secondOut)
	}
}

func TestRedditParse(t *testing.T) {
	const in = "submitted by <a href=\"http://www.reddit.com/user/hexag1\"> hexag1 </a> <br> <a href=\"http://i.imgur.com/toxRCrd.gif\">[link]</a> <a href=\"http://www.reddit.com/r/gifs/comments/1be2d3/forget_the_before_and_after_of_mt_st_helens_heres/\">[190 comments]</a>"
	const firstOut = in
	const secondOut = "http://i.imgur.com/toxRCrd.gif"
	if x, y := RedditParse(in); x != firstOut || y != secondOut {
		t.Errorf("\nTestRedditParse\n(%s)\n = \n(%s,\n%s)\n want \n(%s,\n%s)\n", in, x, y, firstOut, secondOut)
	}
}