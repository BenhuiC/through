package client

import (
	"errors"
	"regexp"
	"strings"
)

var (
	RuleFormatError = errors.New("rule format error")
)

type RuleCondType string

const (
	RuleCondTypeHostMatch  RuleCondType = "host-match"
	RuleCondTypeHostPrefix RuleCondType = "host-prefix"
	RuleCondTypeHostSuffix RuleCondType = "host-suffix"
	RuleCondTypeHostRegexp RuleCondType = "host-regexp"
	RuleCondTypeGEO        RuleCondType = "geo"
	RuleCondTypeIPCIDR     RuleCondType = "ip-cidr"
	RuleCondTypeMatchAll   RuleCondType = "match-all"
)

type RuleActionType string

const (
	RuleActionTypeReject  RuleActionType = "reject"
	RuleActionTypeDirect  RuleActionType = "direct"
	RuleActionTypeForward RuleActionType = "forward"
)

type RuleManager []Rule

func NewRuleManager(rules []string) (r *RuleManager, err error) {
	*r = make([]Rule, 0, len(rules))
	for _, str := range rules {
		ru, err := NewRule(str)
		if err != nil {
			return
		}
		*r = append(*r, ru)
	}

	return
}

func (r *RuleManager) Get(host string) (server string) {
	for _, ru := range *r {
		if ru.Match(host) {
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

	ru := ary[0]
	action := ary[1]

	ruAry := strings.Split(ru, ":")
	r.CondType = RuleCondType(ruAry[0])
	if len(ruAry) == 2 {
		r.CondParam = ruAry[1]
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
		r.Server = acts[1]
	}

	return
}

func (r *Rule) Match(host string) (ok bool) {
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
		// todo
	case RuleCondTypeIPCIDR:
		// todo
		//_, ipnet, _ := net.ParseCIDR(r.CondParam)
		//ok=ipnet.Contains(pv.Resolv(host))
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
