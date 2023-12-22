package v1alpha2

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

const ValidationConfigKey = "validation-config"

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

type validationFunction func(*ValidationConfig) error

func (fn *Function) getBasicValidations() []validationFunction {
	return []validationFunction{
		fn.validateObjectMeta,
	}
}

func (fn *Function) Validate(vc *ValidationConfig) error {
	validations := fn.getBasicValidations()

	if fn.TypeOf(FunctionTypeGit) {
		gitAuthValidators := fn.Spec.gitAuthValidations()
		validations = append(validations, gitAuthValidators...)
	}
	return runValidations(vc, validations...)
}

func runValidations(vc *ValidationConfig, vFuns ...validationFunction) error {
	allErrs := []string{}
	for _, vFun := range vFuns {
		if err := vFun(vc); err != nil {
			allErrs = append(allErrs, err.Error())
		}
	}
	return returnAllErrs("", allErrs)
}

func (fn *Function) validateObjectMeta(_ *ValidationConfig) error {
	nameFn := validation.ValidateNameFunc(validation.NameIsDNS1035Label)
	fieldPath := field.NewPath("metadata")
	if errs := validation.ValidateObjectMeta(&fn.ObjectMeta, true, nameFn, fieldPath); errs != nil {
		return errs.ToAggregate()
	}
	return nil
}

func (spec *FunctionSpec) validateGitRepoURL(_ *ValidationConfig) error {
	if urlIsSSH(spec.Source.GitRepository.URL) {
		return nil
	} else if _, err := url.ParseRequestURI(spec.Source.GitRepository.URL); err != nil {
		return fmt.Errorf("invalid source.gitRepository.URL value: %v", err)
	}
	return nil
}

func (spec *FunctionSpec) gitAuthValidations() []validationFunction {
	if spec.Source.GitRepository.Auth == nil {
		return []validationFunction{}
	}
	return []validationFunction{
		spec.validateGitAuthType,
		spec.validateGitAuthSecretName,
		spec.validateGitRepoURL,
	}
}

func (spec *FunctionSpec) validateGitAuthSecretName(_ *ValidationConfig) error {
	if strings.TrimSpace(spec.Source.GitRepository.Auth.SecretName) == "" {
		return errors.New("spec.source.gitRepository.auth.secretName is required")
	}
	return nil
}

var ErrInvalidGitRepositoryAuthType = fmt.Errorf("invalid git repository authentication type")

func (spec *FunctionSpec) validateGitAuthType(_ *ValidationConfig) error {
	switch spec.Source.GitRepository.Auth.Type {
	case RepositoryAuthBasic, RepositoryAuthSSHKey:
		return nil
	default:
		return ErrInvalidGitRepositoryAuthType
	}
}

func urlIsSSH(repoURL string) bool {
	exp, err := regexp.Compile(`((git|ssh?)|(git@[\w\.]+))(:(//)?)([\w\.@\:/\-~]+)(/)?`)
	if err != nil {
		panic(err)
	}

	return exp.MatchString(repoURL)
}
