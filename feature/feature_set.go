package feature

import (
	"encoding/json"
	"fmt"
	"sort"
	"sync"
)

// Aggregate all available features together.
type FeatureSet struct {
	AESDecrypt      AESDecrypt          `json:"AESDecrypt"`
	EnvInfo         EnvInfo             `json:"EnvInfo"`
	Facebook        Facebook            `json:"Facebook"`
	IMAPAccounts    IMAPAccounts        `json:"IMAPAccounts"`
	SendMail        SendMail            `json:"SendMail"`
	Shell           Shell               `json:"Shell"`
	Twilio          Twilio              `json:"Twilio"`
	Twitter         Twitter             `json:"Twitter"`
	WolframAlpha    WolframAlpha        `json:"WolframAlpha"`
	Undocumented1   Undocumented1       `json:"Undocumented1"`
	LookupByTrigger map[Trigger]Feature `json:"-"`
}

var TestFeatureSet = FeatureSet{} // Features are assigned by init_test.go

// Run initialisation routine on all features, and then populate lookup table for all configured features.
func (fs *FeatureSet) Initialise() error {
	fs.LookupByTrigger = map[Trigger]Feature{}
	triggers := map[Trigger]Feature{
		fs.AESDecrypt.Trigger():    &fs.AESDecrypt,
		fs.EnvInfo.Trigger():       &fs.EnvInfo,
		fs.Facebook.Trigger():      &fs.Facebook,
		fs.IMAPAccounts.Trigger():  &fs.IMAPAccounts,
		fs.SendMail.Trigger():      &fs.SendMail,
		fs.Twilio.Trigger():        &fs.Twilio,
		fs.Shell.Trigger():         &fs.Shell,
		fs.Twitter.Trigger():       &fs.Twitter,
		fs.WolframAlpha.Trigger():  &fs.WolframAlpha,
		fs.Undocumented1.Trigger(): &fs.Undocumented1,
	}
	for trigger, featureRef := range triggers {
		if featureRef.IsConfigured() {
			if err := featureRef.Initialise(); err != nil {
				return err
			}
			fs.LookupByTrigger[trigger] = featureRef
		}
	}
	return nil
}

// Run self test of all configured features in parallel. Return test errors if any.
func (fs *FeatureSet) SelfTest() (ret map[Trigger]error) {
	ret = make(map[Trigger]error)
	retMutex := &sync.Mutex{}
	wait := &sync.WaitGroup{}
	wait.Add(len(fs.LookupByTrigger))
	for _, featureRef := range fs.LookupByTrigger {
		go func(ref Feature) {
			err := ref.SelfTest()
			if err != nil {
				retMutex.Lock()
				ret[ref.Trigger()] = err
				retMutex.Unlock()
			}
			wait.Done()
		}(featureRef)
	}
	wait.Wait()
	return
}

// Deserialise feature configuration from JSON configuration. The function does not initialise features automatically.
func (fs *FeatureSet) DeserialiseFromJSON(configJSON json.RawMessage) error {
	// Turn input JSON into map[string]json.RawMessage, map key is the feature key in JSON.
	var configMap map[string]json.RawMessage
	if err := json.Unmarshal(configJSON, &configMap); err != nil {
		return fmt.Errorf("FeatureSet.DeserialiseFromJSON: failed to retrieve config map - %v", err)
	}
	// Here are the feature keys
	features := map[string]Feature{
		"AESDecrypt":    &fs.AESDecrypt,
		"EnvInfo":       &fs.EnvInfo,
		"Facebook":      &fs.Facebook,
		"IMAPAccounts":  &fs.IMAPAccounts,
		"SendMail":      &fs.SendMail,
		"Shell":         &fs.Shell,
		"Twilio":        &fs.Twilio,
		"Twitter":       &fs.Twitter,
		"WolframAlpha":  &fs.WolframAlpha,
		"Undocumented1": &fs.Undocumented1,
	}
	for featureKey, featureRef := range features {
		if featureJSON, exists := configMap[featureKey]; exists {
			if err := json.Unmarshal(featureJSON, &featureRef); err != nil {
				return fmt.Errorf("FeatureSet.DeserialiseFromJSON: failed to deserialise JSON key %s - %v", featureKey, err)
			}
		}
	}
	return nil
}

// Return all configured & initialised triggers, sorted in alphabetical order.
func (fs *FeatureSet) GetTriggers() []string {
	ret := make([]string, 0, 8)
	if fs.LookupByTrigger == nil {
		return ret
	}
	for trigger, _ := range fs.LookupByTrigger {
		ret = append(ret, string(trigger))
	}
	sort.Strings(ret)
	return ret
}
