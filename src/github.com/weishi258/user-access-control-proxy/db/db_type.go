package db

import (
	"regexp"
	"strings"
)

func DecPermission(permission int8) (ret ProxyPermission){
	ret.Get = (permission & 0x01) == 0x01
	ret.Post = (permission & 0x02) == 0x02
	ret.Put = (permission & 0x04) == 0x04
	ret.Delete = (permission & 0x08) == 0x08
	return ret
}
func EncPermission(input ProxyPermission) (ret int8){
	ret = 0
	if input.Get {
		ret |= 0x01
	}
	if input.Post {
		ret |= 0x02
	}
	if input.Put {
		ret |= 0x04
	}
	if input.Delete {
		ret |= 0x08
	}
	return ret
}

type ProxyPermission struct{
	Get bool
	Post bool
	Put bool
	Delete bool
}


type Group struct{
	Id 		int
	Name 	string
	Desc 	string
	Mutable int
}

func (c *Group)IsMutable() bool{
	return c.Mutable != 0
}
func (c *Group)SetMutable(mutable bool) {
	if mutable{
		c.Mutable = 1
	}else{
		c.Mutable = 0
	}
}


type Rule struct{
	Id 			int
	GroupId		int
	Rule		string
	Permission  int8
	Proxy       string
	Weight 		int
	RuleRegex	*regexp.Regexp
}
func (c * Rule)GetRegex() *regexp.Regexp{
	if c.RuleRegex == nil{
		c.RuleRegex, _ = regexp.Compile(c.Rule)
	}
	return c.RuleRegex
}
func (c *Rule)Match(url string) (ret bool){
	regex := c.GetRegex()
	return regex.MatchString(url)
}

func (c *Rule)IsRemote() bool{
	ret := strings.HasPrefix(c.Proxy, "http://")
	if !ret {
		ret = strings.HasPrefix(c.Proxy,"https://")
	}
	return ret
}
func (c * Rule)ComposeProxyUrl(url string) string {
	return c.Proxy + url
}
func (c *Rule)GetPermission() ProxyPermission{
	return DecPermission(c.Permission)
}
func (c *Rule)SetPermission(permission ProxyPermission){
	c.Permission = EncPermission(permission)
}