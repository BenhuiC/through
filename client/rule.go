package client

import (
	"errors"
	"net"
	"regexp"
	"strings"
)

var (
	RuleFormatError = errors.New("rule format error")
)

type RuleCondType string

const (
	RuleCondTypeHostMatch  RuleCondType = "host-match"  // 字符串包含
	RuleCondTypeHostPrefix RuleCondType = "host-prefix" // 字符串前缀
	RuleCondTypeHostSuffix RuleCondType = "host-suffix" // 字符串后缀
	RuleCondTypeHostRegexp RuleCondType = "host-regexp" // 正则匹配
	RuleCondTypeGEO        RuleCondType = "geo"         // geo地址匹配
	RuleCondTypeIPCIDR     RuleCondType = "ip-cidr"     // cidr匹配
	RuleCondTypeMatchAll   RuleCondType = "match-all"   // always return true
)

type RuleActionType string

const (
	RuleActionTypeReject  RuleActionType = "reject"  // reject request
	RuleActionTypeDirect  RuleActionType = "direct"  // call at local
	RuleActionTypeForward RuleActionType = "forward" // forward to through server
)

type RuleManager struct {
	rules     []Rule
	resolvers *ResolverManager
}

func NewRuleManager(resolvers *ResolverManager, rules []string) (r *RuleManager, err error) {
	r = &RuleManager{
		rules:     make([]Rule, 0, len(rules)),
		resolvers: resolvers,
	}
	for _, str := range rules {
		ru, err := NewRule(str)
		if err != nil {
			return nil, err
		}
		r.rules = append(r.rules, ru)
	}

	return
}

func (r *RuleManager) Get(host string) (server string) {
	if strings.Contains(host, ":") {
		ary := strings.Split(host, ":")
		host = ary[0]
	}
	for _, ru := range r.rules {
		if ru.Match(r.resolvers, host) {
			server = ru.Server
			return
		}
	}
	return
}

type Rule struct {
	CondType  RuleCondType
	Action    RuleActionType
	CondParam string
	Server    string
}

func NewRule(s string) (r Rule, err error) {
	ary := strings.Split(s, ",")
	if len(ary) != 2 {
		err = RuleFormatError
		return
	}

	ru := strings.TrimSpace(ary[0])
	action := strings.TrimSpace(ary[1])

	ruAry := strings.Split(ru, ":")
	r.CondType = RuleCondType(ruAry[0])
	if len(ruAry) == 2 {
		r.CondParam = strings.TrimSpace(ruAry[1])
	}
	if !isLegalRuleCondType(r.CondType) {
		err = RuleFormatError
		return
	}

	if strings.HasPrefix(action, string(RuleActionTypeReject)) {
		r.Action = RuleActionTypeReject
		r.Server = string(RuleActionTypeReject)
	} else if strings.HasPrefix(action, string(RuleActionTypeDirect)) {
		r.Action = RuleActionTypeDirect
		r.Server = string(RuleActionTypeDirect)
	} else if strings.HasPrefix(action, string(RuleActionTypeForward)) {
		r.Action = RuleActionTypeForward
		acts := strings.Split(action, ":")
		if len(acts) != 2 {
			err = RuleFormatError
			return
		}
		r.Server = strings.TrimSpace(acts[1])
	} else {
		err = RuleFormatError
		return
	}

	return
}

func (r *Rule) Match(rs *ResolverManager, host string) (ok bool) {
	switch r.CondType {
	case RuleCondTypeHostMatch:
		ok = strings.Contains(host, r.CondParam)
	case RuleCondTypeHostPrefix:
		ok = strings.HasPrefix(host, r.CondParam)
	case RuleCondTypeHostSuffix:
		ok = strings.HasSuffix(host, r.CondParam)
	case RuleCondTypeHostRegexp:
		reg := regexp.MustCompile(r.CondParam)
		ok = reg.Match([]byte(host))
	case RuleCondTypeGEO:
		ct := rs.Country(host)
		if ct == r.CondParam {
			ok = true
		}
	case RuleCondTypeIPCIDR:
		_, ipnet, _ := net.ParseCIDR(r.CondParam)
		ok = ipnet.Contains(rs.Lookup(host))
	case RuleCondTypeMatchAll:
		ok = true
	}
	return
}

func isLegalRuleCondType(c RuleCondType) (ok bool) {
	switch c {
	case RuleCondTypeHostMatch, RuleCondTypeHostPrefix, RuleCondTypeHostSuffix,
		RuleCondTypeHostRegexp, RuleCondTypeGEO, RuleCondTypeIPCIDR,
		RuleCondTypeMatchAll:
		ok = true
	}

	return
}
