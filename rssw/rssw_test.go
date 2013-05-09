package rssw

import (
	"fmt"
	"testing"
)

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

func TestReturesParse(t *testing.T) {
	const in = "<div class=\"description\">\n" +
		"  CARACAS (Reuters) - Venezuelans went to the polls on Sunday to vote whether to honor Hugo Chavez's dying wish for a longtime loyalist to continue his self-proclaimed socialist revolution or hand power to a young challenger vowing business-friendly changes.<img width=\"1\" height=\"1\" src=\"http://reuters.us.feedsportal.com/c/35217/f/654200/s/2aafbd78/mf.gif\" border=\"0\"><br><br><a href=\"http://da.feedsportal.com/r/163287437350/u/49/f/654200/c/35217/s/2aafbd78/a2.htm\"><img src=\"http://da.feedsportal.com/r/163287437350/u/49/f/654200/c/35217/s/2aafbd78/a2.img\" border=\"0\"></a><img width=\"1\" height=\"1\" src=\"http://pi.feedsportal.com/r/163287437350/u/49/f/654200/c/35217/s/2aafbd78/a2t.img\" border=\"0\"><div class=\"feedflare\">\n" +
		"<a href=\"http://feeds.reuters.com/~ff/reuters/topNews?a=fSgtwfJ9rv0:DzNt6rV3-vw:yIl2AUoC8zA\"><img src=\"http://feeds.feedburner.com/~ff/reuters/topNews?d=yIl2AUoC8zA\" border=\"0\"></a> <a href=\"http://feeds.reuters.com/~ff/reuters/topNews?a=fSgtwfJ9rv0:DzNt6rV3-vw:V_sGLiPBpWU\"><img src=\"http://feeds.feedburner.com/~ff/reuters/topNews?i=fSgtwfJ9rv0:DzNt6rV3-vw:V_sGLiPBpWU\" border=\"0\"></a> <a href=\"http://feeds.reuters.com/~ff/reuters/topNews?a=fSgtwfJ9rv0:DzNt6rV3-vw:-BTjWOF_DHI\"><img src=\"http://feeds.feedburner.com/~ff/reuters/topNews?i=fSgtwfJ9rv0:DzNt6rV3-vw:-BTjWOF_DHI\" border=\"0\"></a>\n" +
		"</div><img src=\"http://feeds.feedburner.com/~r/reuters/topNews/~4/fSgtwfJ9rv0\" height=\"1\" width=\"1\">\n" +
		"</div>"
	const out = "CARACAS (Reuters) - Venezuelans went to the polls on Sunday to vote whether to honor Hugo Chavez's dying wish for a longtime loyalist to continue his self-proclaimed socialist revolution or hand power to a young challenger vowing business-friendly changes."
	i := ItemObject{Description: in}
	channel := make(chan int)
	go ReutersParse(channel, &i)
	<-channel
	if i.Description != out {
		t.Errorf("\nTestReutersParse\n(%s)\n = \n(%s)\n want \n(%s)", in, i.Description, out)
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

func TestSimilarImages(t *testing.T) {
	const first = "Sveriges Dick Axelsson missar en straff"
	const second = "Alex Ferguson och David Moyes tillsammans förra året"
	const title = "David Moyes slutar i Everton"
	score1 := GetSimilarityScore(title, first)
	score2 := GetSimilarityScore(title, second)

	fmt.Printf("%d vs %d\n", score1, score2)
}
