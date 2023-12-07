package v1alpha2

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/validation"
	v1validation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	utilvalidation "k8s.io/apimachinery/pkg/util/validation"
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
		fn.Spec.validateLabels,
		fn.Spec.validateAnnotations,
		fn.Spec.validateSources,
		fn.Spec.validateSecretMounts,
	}
}

var (
	ErrUnknownFunctionType = fmt.Errorf("unknown function source type")
)

func (fn *Function) Validate(vc *ValidationConfig) error {
	validations := fn.getBasicValidations()

	switch {
	case fn.TypeOf(FunctionTypeInline):
		validations = append(validations, fn.Spec.validateInlineSrc, fn.Spec.validateInlineDeps)
		return runValidations(vc, validations...)

	case fn.TypeOf(FunctionTypeGit):
		gitAuthValidators := fn.Spec.gitAuthValidations()
		validations = append(validations, gitAuthValidators...)
		return runValidations(vc, validations...)

	default:
		validations = append(validations, unknownFunctionTypeValidator)
		return runValidations(vc, validations...)
	}
}

func unknownFunctionTypeValidator(_ *ValidationConfig) error {
	return ErrUnknownFunctionType
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

func (spec *FunctionSpec) validateInlineSrc(_ *ValidationConfig) error {
	if spec.Source.Inline.Source == "" {
		return fmt.Errorf("empty source.inline.source value")
	}
	return nil
}

func (spec *FunctionSpec) validateInlineDeps(_ *ValidationConfig) error {
	if err := ValidateDependencies(spec.Runtime, spec.Source.Inline.Dependencies); err != nil {
		return errors.Wrap(err, "invalid source.inline.dependencies value")
	}
	return nil
}

func (spec *FunctionSpec) gitAuthValidations() []validationFunction {
	if spec.Source.GitRepository.Auth == nil {
		return []validationFunction{
			spec.validateRepository,
		}
	}
	return []validationFunction{
		spec.validateRepository,
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

func (spec *FunctionSpec) validateSources(vc *ValidationConfig) error {
	sources := 0
	if spec.Source.GitRepository != nil {
		sources++
	}

	if spec.Source.Inline != nil {
		sources++
	}
	if sources == 1 {
		return nil
	}
	return errors.Errorf("spec.source should contains only 1 configuration of function")
}

func (spec *FunctionSpec) validateLabels(_ *ValidationConfig) error {
	errs := field.ErrorList{}
	errs = append(errs, validateFunctionLabels(spec.Labels, "spec.labels")...)

	return errs.ToAggregate()
}

func validateFunctionLabels(labels map[string]string, path string) field.ErrorList {
	errs := field.ErrorList{}

	fieldPath := field.NewPath(path)
	errs = append(errs, v1validation.ValidateLabels(labels, fieldPath)...)
	errs = append(errs, validateFunctionLabelsByOwnGroup(labels, fieldPath)...)

	return errs
}

func validateFunctionLabelsByOwnGroup(labels map[string]string, fieldPath *field.Path) field.ErrorList {
	forbiddenPrefix := FunctionGroup + "/"
	errorMessage := fmt.Sprintf("label from domain %s is not allowed", FunctionGroup)
	allErrs := field.ErrorList{}
	for k := range labels {
		if strings.HasPrefix(k, forbiddenPrefix) {
			allErrs = append(allErrs, field.Invalid(fieldPath, k, errorMessage))
		}
	}
	return allErrs
}

func (spec *FunctionSpec) validateAnnotations(_ *ValidationConfig) error {
	fieldPath := field.NewPath("spec.annotations")
	errs := validation.ValidateAnnotations(spec.Annotations, fieldPath)

	return errs.ToAggregate()
}

func (spec *FunctionSpec) validateSecretMounts(_ *ValidationConfig) error {
	var allErrs []string
	secretMounts := spec.SecretMounts
	for _, secretMount := range secretMounts {
		allErrs = append(allErrs,
			utilvalidation.IsDNS1123Subdomain(secretMount.SecretName)...)
	}

	if !secretNamesAreUnique(secretMounts) {
		allErrs = append(allErrs, "secretNames should be unique")
	}

	if !secretMountPathAreNotEmpty(secretMounts) {
		allErrs = append(allErrs, "mountPath should not be empty")
	}

	return returnAllErrs("invalid spec.secretMounts", allErrs)
}

func secretNamesAreUnique(secretMounts []SecretMount) bool {
	uniqueSecretNames := make(map[string]bool)
	for _, secretMount := range secretMounts {
		uniqueSecretNames[secretMount.SecretName] = true
	}
	return len(uniqueSecretNames) == len(secretMounts)
}

func secretMountPathAreNotEmpty(secretMounts []SecretMount) bool {
	for _, secretMount := range secretMounts {
		if secretMount.MountPath == "" {
			return false
		}
	}
	return true
}

type property struct {
	name  string
	value string
}

func (spec *FunctionSpec) validateRepository(_ *ValidationConfig) error {
	return validateIfMissingFields([]property{
		{name: "spec.source.gitRepository.baseDir", value: spec.Source.GitRepository.BaseDir},
		{name: "spec.source.gitRepository.reference", value: spec.Source.GitRepository.Reference},
	}...)
}

func urlIsSSH(repoURL string) bool {
	exp, err := regexp.Compile(`((git|ssh?)|(git@[\w\.]+))(:(//)?)([\w\.@\:/\-~]+)(/)?`)
	if err != nil {
		panic(err)
	}

	return exp.MatchString(repoURL)
}
