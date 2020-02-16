package generic

import (
	"crypto/hmac"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"

	"github.com/golang/glog"

	kapi "k8s.io/kubernetes/pkg/api"

	"github.com/openshift/origin/pkg/build/api"
	"github.com/openshift/origin/pkg/build/webhook"
)

// WebHookPlugin used for processing manual(or other) webhook requests.
type WebHookPlugin struct{}

// New returns a generic webhook plugin.
func New() *WebHookPlugin {
	return &WebHookPlugin{}
}

// Extract services generic webhooks.
func (p *WebHookPlugin) Extract(buildCfg *api.BuildConfig, secret, path string, req *http.Request) (revision *api.SourceRevision, envvars []kapi.EnvVar, proceed bool, err error) {
	trigger, ok := webhook.FindTriggerPolicy(api.GenericWebHookBuildTriggerType, buildCfg)
	if !ok {
		err = webhook.ErrHookNotEnabled
		return
	}
	glog.V(4).Infof("Checking if the provided secret for BuildConfig %s/%s matches", buildCfg.Namespace, buildCfg.Name)
	if !hmac.Equal([]byte(trigger.GenericWebHook.Secret), []byte(secret)) {
		err = webhook.ErrSecretMismatch
		return
	}
	glog.V(4).Infof("Verifying build request for BuildConfig %s/%s", buildCfg.Namespace, buildCfg.Name)
	if err = verifyRequest(req); err != nil {
		return
	}

	git := buildCfg.Spec.Source.Git
	if git == nil {
		glog.V(4).Infof("No source defined for BuildConfig %s/%s, but triggering anyway", buildCfg.Namespace, buildCfg.Name)
		return nil, envvars, true, nil
	}

	contentType := req.Header.Get("Content-Type")
	if len(contentType) != 0 {
		contentType, _, err = mime.ParseMediaType(contentType)
		if err != nil {
			return nil, envvars, false, fmt.Errorf("non-parseable Content-Type %s (%s)", contentType, err)
		}
	}

	if req.Body != nil && contentType == "application/json" {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, envvars, false, err
		}
		if len(body) == 0 {
			return nil, envvars, true, nil
		}
		var data api.GenericWebHookEvent
		if err = json.Unmarshal(body, &data); err != nil {
			glog.V(4).Infof("Error unmarshaling json %v, but continuing", err)
			return nil, envvars, true, nil
		}
		if len(data.Env) > 0 && trigger.GenericWebHook.AllowEnv {
			envvars = data.Env
		}
		if data.Git == nil {
			glog.V(4).Infof("No git information for the generic webhook found in %s/%s", buildCfg.Namespace, buildCfg.Name)
			return nil, envvars, true, nil
		}

		if data.Git.Refs != nil {
			for _, ref := range data.Git.Refs {
				if webhook.GitRefMatches(ref.Ref, git.Ref) {
					revision = &api.SourceRevision{
						Git: &ref.GitSourceRevision,
					}
					return revision, envvars, true, nil
				}
			}
			glog.V(2).Infof("Skipping build for BuildConfig %s/%s. None of the supplied refs matched %q", buildCfg.Namespace, buildCfg, git.Ref)
			return nil, envvars, false, nil
		}
		if !webhook.GitRefMatches(data.Git.Ref, git.Ref) {
			glog.V(2).Infof("Skipping build for BuildConfig %s/%s. Branch reference from %q does not match configuration", buildCfg.Namespace, buildCfg.Name, data.Git.Ref)
			return nil, envvars, false, nil
		}
		revision = &api.SourceRevision{
			Git: &data.Git.GitSourceRevision,
		}
	}
	return revision, envvars, true, nil
}

func verifyRequest(req *http.Request) error {
	if method := req.Method; method != "POST" {
		return fmt.Errorf("Unsupported HTTP method %s", method)
	}
	return nil
}
