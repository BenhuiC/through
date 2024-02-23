package client

type RuleCondType string

const (
	RuleCondType_HostMatch  RuleCondType = "host-match"
	RuleCondType_HostPrefix RuleCondType = "host-prefix"
	RuleCondType_HostSuffix RuleCondType = "host-suffix"
	RuleCondType_HostRegexp RuleCondType = "host-regexp"
	RuleCondType_GEOIP      RuleCondType = "geoip"
	RuleCondType_IPCIDR     RuleCondType = "ip-cidr"
	RuleCondType_HasServer  RuleCondType = "has-server"
	RuleCondType_MatchAll   RuleCondType = "match-all"
)

type RuleActionType string

const (
	RuleActionType_Reject  RuleActionType = "reject"
	RuleActionType_Direct  RuleActionType = "direct"
	RuleActionType_Forward RuleActionType = "forward"
)

type RuleManager struct {
	forwardClients []RuleForward
}

func NewRuleManager(rules []string) (r *RuleManager, err error) {
	// todo

	//r = &RuleManager{forwardClients: make([]RuleForward, 0)}
	//for _, str := range rules {
	//	ru, err := NewRule(str)
	//	if err != nil {
	//		return
	//	}
	//	r.forwardClients = append(r.forwardClients, ru)
	//}

	return
}

func (r *RuleManager) Get(host string) (f Forward) {
	for i, ru := range r.forwardClients {
		if ru.Match(host) {
			f = r.forwardClients[i]
			return
		}
	}
	return
}

type Rule struct {
	CondType RuleCondType
	Action   RuleActionType
	Param    string
}

func NewRule(s string) (r Rule, err error) {
	// todo
	return
}

func (r *Rule) Match(host string) bool {
	// todo
	return true
}

type RuleForward struct {
	Rule
	Forward
}
