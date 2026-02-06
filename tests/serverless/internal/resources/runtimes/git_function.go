package runtimes

import (
	serverlessv1alpha2 "github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
)

type GitopsFunctionBuilder struct {
	repoURL         string
	baseDir         string
	reference       string
	runtime         serverlessv1alpha2.Runtime
	auth            *serverlessv1alpha2.RepositoryAuth
	minReplicas     int32
	maxReplicas     int32
	functionProfile string
	buildProfile    string
	labels          map[string]string
}

func NewGitopsFunctionBuilder(repoURL string, runtime serverlessv1alpha2.Runtime) *GitopsFunctionBuilder {
	return &GitopsFunctionBuilder{
		repoURL:         repoURL,
		baseDir:         "/",
		reference:       "main",
		runtime:         runtime,
		minReplicas:     1,
		maxReplicas:     2,
		functionProfile: "M",
		buildProfile:    "fast",
		labels:          make(map[string]string),
	}
}

func (b *GitopsFunctionBuilder) BaseDir(baseDir string) *GitopsFunctionBuilder {
	if baseDir != "" {
		b.baseDir = baseDir
	}

	return b
}

func (b *GitopsFunctionBuilder) Reference(reference string) *GitopsFunctionBuilder {
	if reference != "" {
		b.reference = reference
	}

	return b
}

func (b *GitopsFunctionBuilder) Auth(auth *serverlessv1alpha2.RepositoryAuth) *GitopsFunctionBuilder {
	b.auth = auth
	return b
}

func (b *GitopsFunctionBuilder) MinReplicas(minReplicas int32) *GitopsFunctionBuilder {
	b.minReplicas = minReplicas
	return b
}

func (b *GitopsFunctionBuilder) MaxReplicas(maxReplicas int32) *GitopsFunctionBuilder {
	b.maxReplicas = maxReplicas
	return b
}

func (b *GitopsFunctionBuilder) FunctionProfile(functionProfile string) *GitopsFunctionBuilder {
	if functionProfile != "" {
		b.functionProfile = functionProfile
	}
	return b
}

func (b *GitopsFunctionBuilder) BuildProfile(buildProfile string) *GitopsFunctionBuilder {
	if buildProfile != "" {
		b.buildProfile = buildProfile
	}
	return b
}

func (b *GitopsFunctionBuilder) AddLabel(key, value string) *GitopsFunctionBuilder {
	b.labels[key] = value

	return b
}

func (b *GitopsFunctionBuilder) Build() serverlessv1alpha2.FunctionSpec {
	gitRepo := &serverlessv1alpha2.GitRepositorySource{
		URL: b.repoURL,
		Repository: serverlessv1alpha2.Repository{
			BaseDir:   b.baseDir,
			Reference: b.reference,
		},
	}
	if b.auth != nil {
		gitRepo.Auth = b.auth
	}
	return serverlessv1alpha2.FunctionSpec{
		Runtime: b.runtime,
		Source: serverlessv1alpha2.Source{
			GitRepository: gitRepo,
		},
		ScaleConfig: &serverlessv1alpha2.ScaleConfig{
			MinReplicas: &b.minReplicas,
			MaxReplicas: &b.maxReplicas,
		},
		ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
			Function: &serverlessv1alpha2.ResourceRequirements{
				Profile: b.functionProfile,
			},
			Build: &serverlessv1alpha2.ResourceRequirements{
				Profile: b.buildProfile,
			},
		},
		Labels: b.labels,
	}
}
