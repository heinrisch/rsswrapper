package rssw

import "testing"

func TestAftonbladetParseFull(t *testing.T) {
	const in = "<![CDATA[<img src=\"http://gfx.aftonbladet-cdn.se/image/16518937/250/widescreen/49db9ec601a95/Ger+sig+ut+p%C3%A5+ov%C3%A4ntat+djupt+vatten \" /><p>Se när han grundluras av polarna</p>]]>"
	const firstOut = "Se när han grundluras av polarna"
	const secondOut = "http://gfx.aftonbladet-cdn.se/image/16518937/250/widescreen/49db9ec601a95/Ger+sig+ut+p%C3%A5+ov%C3%A4ntat+djupt+vatten"
	i := ItemObject{Description: in}
	channel := make(chan int)
	go AftonbladetParse(channel, &i)
	<-channel
	if i.Description != firstOut || i.ParsedImage != secondOut {
		t.Errorf("\nTestAftonbladetParse\n(%s)\n = \n(%s,\n%s)\n want \n(%s,\n%s)", in, i.Description, i.ParsedImage, firstOut, secondOut)
	}
}

func TestAftonbladetParseEmpty(t *testing.T) {
	const in = "<![CDATA[<p>Peter Kadhammars berättelse om ett Sverige som varit, och ett som blir. • För att läsa: Ladda ner dokumentet som pdf eller zooma in direkt på sidan.</p>]]>"
	const firstOut = "Peter Kadhammars berättelse om ett Sverige som varit, och ett som blir. • För att läsa: Ladda ner dokumentet som pdf eller zooma in direkt på sidan."
	const secondOut = ""
	i := ItemObject{}
	i.Description = in
	channel := make(chan int)
	go AftonbladetParse(channel, &i)
	<-channel
	if i.Description != firstOut || i.ParsedImage != secondOut {
		t.Errorf("\nTestAftonbladetParse\n(%s)\n = \n(%s,\n%s)\n want \n(%s,\n%s)\n", in, i.Description, i.ParsedImage, firstOut, secondOut)
	}
}

func TestRedditParse(t *testing.T) {
	const in = "submitted by <a href=\"http://www.reddit.com/user/hexag1\"> hexag1 </a> <br> <a href=\"http://i.imgur.com/toxRCrd.gif\">[link]</a> <a href=\"http://www.reddit.com/r/gifs/comments/1be2d3/forget_the_before_and_after_of_mt_st_helens_heres/\">[190 comments]</a>"
	const firstOut = in
	const secondOut = "http://i.imgur.com/toxRCrd.gif"
	i := ItemObject{}
	i.Description = in
	channel := make(chan int)
	go RedditParse(channel, &i)
	<-channel
	if i.Description != firstOut || i.ParsedImage != secondOut {
		t.Errorf("\nTestRedditParse\n(%s)\n = \n(%s,\n%s)\n want \n(%s,\n%s)\n", in, i.Description, i.ParsedImage, firstOut, secondOut)
	}
}

//This is the worst slowest test ever...
func TestMetaParse(t *testing.T) {
	const in = "http://www.svd.se/nyheter/inrikes/tolkorganisationer-pressar-reinfeldt_8051094.svd" //hopefully this article will be alive for a while...
	const firstOut = ""
	const secondOut = "http://gfx.svd-cdn.se/multimedia/dynamic/01010/brp-Afghanistan-to_1010324c.jpg"
	i := ItemObject{}
	i.Link = in
	channel := make(chan int)
	go MetaParse(channel, &i)
	<-channel
	if i.Description != firstOut || i.ParsedImage != secondOut {
		t.Errorf("\nTestMetaParse\n(%s)\n = \n(%s,\n%s)\n want \n(%s,\n%s)\n", in, i.Description, i.ParsedImage, firstOut, secondOut)
	}
}

func TestMetaParse2(t *testing.T) {
	const in = "http://www.dn.se/nyheter/varlden/usa-bygger-upp-raketforsvar-pa-guam" //hopefully this article will be alive for a while...
	const firstOut = ""
	const secondOut = "http://www.dn.se/Images/2013/04/03/nordkoreaGuam683.jpg"
	i := ItemObject{}
	i.Link = in
	channel := make(chan int)
	go MetaParse(channel, &i)
	<-channel
	if i.Description != firstOut || i.ParsedImage != secondOut {
		t.Errorf("\nTestMetaParse2\n(%s)\n = \n(%s,\n%s)\n want \n(%s,\n%s)\n", in, i.Description, i.ParsedImage, firstOut, secondOut)
	}
}
