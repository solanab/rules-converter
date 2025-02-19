package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/sagernet/sing-box/common/srs"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	flagMixMode       bool
	flagConvertOutput string
	flagVersion       uint8
)

const (
	flagConvertDefaultOutput = "<file_name>"
	defaultVersion          = 3
)

var mainCommand = &cobra.Command{
	Use:   "sing-rules-converter [source-path]",
	Short: "convert clash, surge rule-provider to sing-box rule-set format",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := compileRuleSet(args[0])
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	mainCommand.Flags().StringVarP(&flagConvertOutput, "output", "o", flagConvertDefaultOutput, "Output file name")
	mainCommand.Flags().BoolVarP(&flagMixMode, "mix", "m", false, "Mix mode to combine different rule types")
	mainCommand.Flags().Uint8VarP(&flagVersion, "version", "v", defaultVersion, "Rule set version (1-3)")
}

func main() {
	if err := mainCommand.Execute(); err != nil {
		log.Fatal(err)
	}
}

type clashRuleProvider struct {
	Rules []string `yaml:"payload,omitempty"`
}

func getRulesFromContent(content []byte) ([]string, error) {
	var provider clashRuleProvider
	if err := yaml.Unmarshal(content, &provider); err == nil {
		return provider.Rules, nil
	}
	
	var ruleArr []string
	for _, lineRaw := range strings.Split(string(content), "\n") {
		line := strings.TrimSpace(lineRaw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if !strings.Contains(line, ",") {
			continue
		}
		ruleArr = append(ruleArr, line)
	}
	return ruleArr, nil
}

func saveSourceRuleSet(ruleset *option.PlainRuleSetCompat, outputPath string) error {
	versionedRuleset := option.PlainRuleSetCompat{
		Version: flagVersion,
		Options: ruleset.Options,
	}
	
	buffer := new(bytes.Buffer)
	encoder := json.NewEncoder(buffer)
	encoder.SetIndent("", "  ")
	
	if err := encoder.Encode(&versionedRuleset); err != nil {
		return E.Cause(err, "encode config")
	}
	
	output, err := os.Create(outputPath)
	if err != nil {
		return E.Cause(err, "open output")
	}
	defer output.Close()
	
	_, err = output.Write(buffer.Bytes())
	if err != nil {
		return E.Cause(err, "write output")
	}
	
	return nil
}

func saveRuleSet(rules []option.DefaultHeadlessRule, outputPath string) error {
	plainRuleSet := option.PlainRuleSetCompat{
		Version: 1,
		Options: option.PlainRuleSet{
			Rules: common.Map(rules, func(it option.DefaultHeadlessRule) option.HeadlessRule {
				return option.HeadlessRule{
					Type:           C.RuleTypeDefault,
					DefaultOptions: it,
				}
			}),
		},
	}

	if err := saveSourceRuleSet(&plainRuleSet, outputPath+"-v"+strconv.FormatUint(uint64(flagVersion), 10)+".json"); err != nil {
		return err
	}

	if err := saveBinaryRuleSet(&plainRuleSet, outputPath+"-v"+strconv.FormatUint(uint64(flagVersion), 10)+".srs"); err != nil {
		return err
	}

	return nil
}

func readYamlAndListToRuleset(content []byte, outputPath string) error {
	rawRules, err := getRulesFromContent(content)
	if err != nil {
		return err
	}
	var (
		rules            []option.DefaultHeadlessRule
		hasStarOnlyRule  bool
		domainArr        []string
		domainSuffixArr  []string
		domainKeywordArr []string
		domainRegexArr   []string
		ipcidrArr        []string
		srcipcidrArr     []string
		portArr          []uint16
		srcPortArr       []uint16
		processNameArr   []string
		processPathArr   []string
	)
	for _, line := range rawRules {
		if strings.Contains(line, "AND") || strings.Contains(line, "OR") || strings.Contains(line, "NOT") {
			continue
		}
		strArr := strings.Split(line, ",")
		ruleType := strArr[0]
		ruleContent := strArr[1]
		switch ruleType {
		case "DOMAIN":
			if ruleContent == "*" {
				hasStarOnlyRule = true
				continue
			}
			if strings.Contains(ruleContent, "*") {
				ruleContent = strings.ReplaceAll(ruleContent, "*", "[^\\.]*?")
				domainRegexArr = append(domainRegexArr, ruleContent[1:])
				continue
			}
			if strings.HasPrefix(ruleContent, "+.") {
				domainSuffixArr = append(domainSuffixArr, ruleContent[1:])
				continue
			}
			domainArr = append(domainArr, ruleContent)
		case "DOMAIN-KEYWORD":
			domainKeywordArr = append(domainKeywordArr, ruleContent)
		case "DOMAIN-SUFFIX":
			if ruleContent[0] == '.' {
				domainSuffixArr = append(domainSuffixArr, ruleContent)
				continue
			}
			domainArr = append(domainArr, ruleContent)
			domainSuffixArr = append(domainSuffixArr, "."+ruleContent)
		case "DOMAIN-REGEX":
			domainRegexArr = append(domainRegexArr, ruleContent)
		case "IP-CIDR", "IP-CIDR6":
			ipcidrArr = append(ipcidrArr, ruleContent)
		case "SRC-IP-CIDR":
			srcipcidrArr = append(srcipcidrArr, ruleContent)
		case "DST-PORT":
			port, _ := strconv.Atoi(ruleContent)
			portArr = append(portArr, uint16(port))
		case "SRC-DST-PORT":
			port, _ := strconv.Atoi(ruleContent)
			srcPortArr = append(srcPortArr, uint16(port))
		case "PROCESS-NAME":
			processNameArr = append(processNameArr, ruleContent)
		case "PROCESS-PATH":
			processPathArr = append(processPathArr, ruleContent)
		}
	}
	if !flagMixMode {
		if len(domainArr) > 0 || len(domainSuffixArr) > 0 || len(domainKeywordArr) > 0 || len(domainRegexArr) > 0 {
			domainRuleArr := []option.DefaultHeadlessRule{
				{
					Domain:        domainArr,
					DomainKeyword: domainKeywordArr,
					DomainSuffix:  domainSuffixArr,
					DomainRegex:   domainRegexArr,
				},
			}
			if hasStarOnlyRule {
				domainRuleArr = append(domainRuleArr, option.DefaultHeadlessRule{
					DomainKeyword: []string{"."},
					Invert:        true,
				})
			}
			if err := saveRuleSet(domainRuleArr, outputPath+"-site"); err != nil {
				return err
			}
			rules = append(rules, domainRuleArr...)
		}
		if len(ipcidrArr) > 0 {
			ipcidrRuleArr := []option.DefaultHeadlessRule{
				{
					IPCIDR: ipcidrArr,
				},
			}
			if err := saveRuleSet(ipcidrRuleArr, outputPath+"-ip"); err != nil {
				return err
			}
			rules = append(rules, ipcidrRuleArr...)
		}
	} else {
		if len(domainArr) > 0 || len(domainSuffixArr) > 0 || len(domainKeywordArr) > 0 || len(domainRegexArr) > 0 || len(ipcidrArr) > 0 {
			rules = append(rules, option.DefaultHeadlessRule{
				Domain:        domainArr,
				DomainKeyword: domainKeywordArr,
				DomainSuffix:  domainSuffixArr,
				DomainRegex:   domainRegexArr,
				IPCIDR:        ipcidrArr,
			})
			if hasStarOnlyRule {
				rules = append(rules, option.DefaultHeadlessRule{
					DomainKeyword: []string{"."},
					Invert:        true,
				})
			}
		}
	}
	if len(portArr) > 0 {
		portRuleArr := []option.DefaultHeadlessRule{
			{
				Port: portArr,
			},
		}
		if !flagMixMode {
			if err := saveRuleSet(portRuleArr, outputPath+"-port"); err != nil {
				return err
			}
		}
		rules = append(rules, portRuleArr...)
	}
	if len(srcPortArr) > 0 {
		srcPortRuleArr := []option.DefaultHeadlessRule{
			{
				SourcePort: srcPortArr,
			},
		}
		if !flagMixMode {
			if err := saveRuleSet(srcPortRuleArr, outputPath+"-src-port"); err != nil {
				return err
			}
		}
		rules = append(rules, srcPortRuleArr...)
	}
	if len(srcipcidrArr) > 0 {
		srcIPCidrRuleArr := []option.DefaultHeadlessRule{
			{
				SourceIPCIDR: srcipcidrArr,
			},
		}
		if !flagMixMode {
			if err := saveRuleSet(srcIPCidrRuleArr, outputPath+"-src-ip"); err != nil {
				return err
			}
		}
		rules = append(rules, srcIPCidrRuleArr...)
	}
	if len(processNameArr) > 0 || len(processPathArr) > 0 {
		var processRuleArr []option.DefaultHeadlessRule
		if len(processNameArr) > 0 {
			processRuleArr = append(processRuleArr, option.DefaultHeadlessRule{
				ProcessName: processNameArr,
			})
		}
		if len(processPathArr) > 0 {
			processRuleArr = append(processRuleArr, option.DefaultHeadlessRule{
				ProcessPath: processPathArr,
			})
		}
		if !flagMixMode {
			if err := saveRuleSet(processRuleArr, outputPath+"-process"); err != nil {
				return err
			}
		}
		rules = append(rules, processRuleArr...)
	}
	if len(processPathArr) > 0 {
		rule := option.DefaultHeadlessRule{
			ProcessPath: processPathArr,
		}
		rules = append(rules, rule)
	}
	if len(rules) == 0 {
		return E.Cause(E.New("empty input"), "no valid rules found in file: "+outputPath)
	}
	if flagMixMode {
		return saveRuleSet(rules, outputPath)
	}
	return nil
}

func saveBinaryRuleSet(ruleset *option.PlainRuleSetCompat, outputPath string) error {
	ruleSet, err := ruleset.Upgrade()
	if err != nil {
		return err
	}
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	err = srs.Write(outputFile, ruleSet, flagVersion)
	if err != nil {
		outputFile.Close()
		os.Remove(outputPath)
		return err
	}
	outputFile.Close()
	return nil
}

func compileRuleSet(sourcePath string) error {
	var (
		reader io.Reader
		err    error
	)
	if sourcePath == "stdin" {
		reader = os.Stdin
	} else {
		file, err := os.Open(sourcePath)
		if err != nil {
			return E.Cause(err, "failed to open source file: "+sourcePath)
		}
		defer file.Close()
		reader = file
	}
	
	content, err := io.ReadAll(reader)
	if err != nil {
		return E.Cause(err, "failed to read content from: "+sourcePath)
	}
	if len(content) == 0 {
		return E.Cause(E.New("empty input"), "file is empty: "+sourcePath)
	}
	
	var outputPath string
	if flagConvertOutput == flagConvertDefaultOutput {
		outputPath = sourcePath
		switch {
		case strings.HasSuffix(sourcePath, ".yaml"):
			outputPath = sourcePath[:len(sourcePath)-5]
		case strings.HasSuffix(sourcePath, ".list"):
			outputPath = sourcePath[:len(sourcePath)-5]
		}
	} else {
		outputPath = flagConvertOutput
	}
	return readYamlAndListToRuleset(content, outputPath)
}
