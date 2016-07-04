package utility

import (
	"fmt"
	"github.com/FidelityInternational/virgil/Godeps/_workspace/src/github.com/cloudfoundry-community/go-cfclient"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// FirewallRules - A collection of Firewall Rules with version
type FirewallRules struct {
	SchemaVersion string         `yaml:"schema_version"`
	FirewallRules []FirewallRule `yaml:"firewall_rules"`
}

// FirewallRule struct
type FirewallRule struct {
	Port        string
	Destination []string
	Protocol    string
	Source      []string
}

// ByPort - implements sort.Interface for []FirewallRule bases on the Port field
type ByPort []FirewallRule

func (p ByPort) Len() int      { return len(p) }
func (p ByPort) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p ByPort) Less(i, j int) bool {
	if strings.EqualFold(p[i].Protocol, "TCP") && strings.EqualFold(p[j].Protocol, "UDP") {
		return true
	} else if strings.EqualFold(p[i].Protocol, "UDP") && strings.EqualFold(p[j].Protocol, "TCP") {
		return false
	}
	portIInt, _ := strconv.Atoi(p[i].Port)
	portJInt, _ := strconv.Atoi(p[j].Port)
	return portIInt < portJInt
}

// PortExpand - serperates port string into array, for example 2,5-7 becomes {2 5 6 7}
func PortExpand(portString string) ([]string, error) {
	var ports []string
	portsBefore := strings.Split(portString, ",")
	for _, port := range portsBefore {
		port = strings.TrimSpace(port)
		if strings.Contains(port, "-") {
			startFinish := strings.Split(port, "-")
			startString := strings.TrimSpace(startFinish[0])
			start, err := strconv.Atoi(startString)
			if err != nil || start <= 0 || start >= 65536 {
				return []string{}, fmt.Errorf("Port %s was invalid as part of range %s", startString, port)
			}
			endString := strings.TrimSpace(startFinish[1])
			end, err := strconv.Atoi(endString)
			if err != nil || end <= 0 || end >= 65536 {
				return []string{}, fmt.Errorf("Port %s was invalid as part of range %s", endString, port)
			}
			if len(startFinish) != 2 || start >= end {
				return []string{}, fmt.Errorf("Port range %s was invalid", port)
			}
			for i := start; i <= end; i++ {
				ports = append(ports, strconv.Itoa(i))
			}
		} else {
			portInt, err := strconv.Atoi(port)
			if err != nil || portInt <= 0 || portInt >= 65536 {
				return []string{}, fmt.Errorf("Port %s was invalid", port)
			}
			ports = append(ports, port)
		}
	}
	return ports, nil
}

// ProcessRule - returns a concise list of firewall rules for one security group rule
func ProcessRule(secGroupRule cfclient.SecGroupRule, firewallRules []FirewallRule, source []string) ([]FirewallRule, error) {
	if strings.EqualFold(secGroupRule.Protocol, "all") {
		newRules := FirewallRule{
			Protocol:    secGroupRule.Protocol,
			Destination: []string{secGroupRule.Destination},
			Source:      source,
		}
		firewallRules = append(firewallRules, newRules)
		return firewallRules, nil
	}
	ports, err := PortExpand(secGroupRule.Ports)
	if err != nil {
		return []FirewallRule{}, err
	}
	for _, port := range ports {
		var newRule = true
		for i, rule := range firewallRules {
			if rule.Port == port && rule.Protocol == secGroupRule.Protocol {
				rule.Destination = append(rule.Destination, secGroupRule.Destination)
				RemoveDuplicates(&rule.Destination)
				firewallRules[i] = rule
				newRule = false
			}
		}
		if newRule {
			newRules := FirewallRule{
				Port:        port,
				Protocol:    secGroupRule.Protocol,
				Destination: []string{secGroupRule.Destination},
				Source:      source,
			}
			firewallRules = append(firewallRules, newRules)
		}
	}
	return firewallRules, nil
}

// RemoveDuplicates - removes duplicated from array of strings
func RemoveDuplicates(xs *[]string) {
	found := make(map[string]bool)
	j := 0
	for i, x := range *xs {
		if !found[x] {
			found[x] = true
			(*xs)[j] = (*xs)[i]
			j++
		}
	}
	*xs = (*xs)[:j]
}

// GetUsedSecGroups - Trims out any security-groups that cannot be used. I.E not running, staging or bound
func GetUsedSecGroups(allSecGroups []cfclient.SecGroup) []cfclient.SecGroup {
	var secGroups []cfclient.SecGroup
	for _, secGroup := range allSecGroups {
		if secGroup.Running || secGroup.Staging || len(secGroup.SpacesData) != 0 {
			secGroups = append(secGroups, secGroup)
		}
	}
	return secGroups
}

// GetFirewallRules - Returns a concise list of firewall rules for all security groups
func GetFirewallRules(source []string, secGroups []cfclient.SecGroup) FirewallRules {
	var (
		firewallRules, fwRules FirewallRules
		err                    error
	)
	firewallRules.SchemaVersion = "1"
	for _, secGroup := range secGroups {
		for _, secGroupRule := range secGroup.Rules {
			fwRules.FirewallRules, err = ProcessRule(secGroupRule, firewallRules.FirewallRules, source)
			if err != nil {
				continue
			}
			firewallRules.FirewallRules = fwRules.FirewallRules
		}
	}
	return compressDuplicateDestinations(firewallRules)
}

func compressDuplicateDestinations(firewallRules FirewallRules) FirewallRules {
	var firewallRulesResult []FirewallRule
	schema := firewallRules.SchemaVersion
	fwRules := firewallRules.FirewallRules
	sort.Sort(ByPort(fwRules))
	for i, fwRule := range fwRules {
		if i == 0 {
			firewallRulesResult = append(firewallRulesResult, fwRule)
			continue
		}
		prevRule := fwRules[i-1]
		if strings.EqualFold(fwRule.Protocol, "ALL") || fwRule.Protocol != prevRule.Protocol {
			firewallRulesResult = append(firewallRulesResult, fwRule)
			continue
		}
		rulePortInt, _ := strconv.Atoi(fwRule.Port)
		prevRulePortInt, _ := strconv.Atoi(prevRule.Port)
		if rulePortInt == prevRulePortInt+1 {
			if reflect.DeepEqual(fwRule.Destination, prevRule.Destination) {
				previousPort := strings.Split(firewallRulesResult[len(firewallRulesResult)-1].Port, "-")[0]
				firewallRulesResult[len(firewallRulesResult)-1].Port = fmt.Sprintf("%s-%s", previousPort, fwRule.Port)
			} else {
				firewallRulesResult = append(firewallRulesResult, fwRule)
			}
		} else {
			firewallRulesResult = append(firewallRulesResult, fwRule)
		}
	}
	return FirewallRules{SchemaVersion: schema, FirewallRules: firewallRulesResult}
}
