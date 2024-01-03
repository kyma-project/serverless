package v1alpha2

import (
	"fmt"
	"net/url"
	"regexp"
)

type MinFunctionResourcesValues struct {
	MinRequestCPU    string
	MinRequestMemory string
}

type MinBuildJobResourcesValues struct {
	MinRequestCPU    string
	MinRequestMemory string
}

type MinFunctionValues struct {
	Resources MinFunctionResourcesValues
}

type MinBuildJobValues struct {
	Resources MinBuildJobResourcesValues
}

type ValidationConfig struct {
	ReservedEnvs []string
	Function     MinFunctionValues
	BuildJob     MinBuildJobValues
}

func urlIsSSH(repoURL string) bool {
	exp, err := regexp.Compile(`((git|ssh)|(git@[\w\.]+))(:(//)?)([\w\.@\:/\-~]+)(/)?`)
	if err != nil {
		panic(err)
	}

	return exp.MatchString(repoURL)
}

func ValidateGitRepoURL(gitRepo *GitRepositorySource) error {
	if urlIsSSH(gitRepo.URL) {
		return nil
	} else if _, err := url.ParseRequestURI(gitRepo.URL); err != nil {
		return fmt.Errorf("source.gitRepository.URL: %v", err)
	}
	return nil
}
