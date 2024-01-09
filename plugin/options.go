package plugin

import "time"

// WithServer configures with a custom SCM server
func WithServer(server string) func(*Plugin) {
	return func(p *Plugin) {
		p.server = server
	}
}

// WithGithubToken configures with the github token specified
func WithGithubToken(gitHubToken string) func(*Plugin) {
	return func(p *Plugin) {
		p.gitHubToken = gitHubToken
	}
}

// WithGitlabToken configures with the gitlab token specified
func WithGitlabToken(gitLabToken string) func(*Plugin) {
	return func(p *Plugin) {
		p.gitLabToken = gitLabToken
	}
}

// WithGitlabServer configures with the gitlab server specified
func WithGitlabServer(gitLabServer string) func(*Plugin) {
	return func(p *Plugin) {
		p.gitLabServer = gitLabServer
	}
}

// WithBitBucketAuthServer configures an auth server
func WithBitBucketAuthServer(bitBucketAuthServer string) func(*Plugin) {
	return func(p *Plugin) {
		p.bitBucketAuthServer = bitBucketAuthServer
	}
}

// WithBitBucketClient configures with a bitbucket client, alternative to github
func WithBitBucketClient(bitBucketClient string) func(*Plugin) {
	return func(p *Plugin) {
		p.bitBucketClient = bitBucketClient
	}
}

// WithBitBucketClient configures with a bitbucket secret, alternative to github
func WithBitBucketSecret(bitBucketSecret string) func(*Plugin) {
	return func(p *Plugin) {
		p.bitBucketSecret = bitBucketSecret
	}
}

// WithConcat configures with concat enabled or disabled
func WithConcat(concat bool) func(*Plugin) {
	return func(p *Plugin) {
		p.concat = concat
	}
}

// WithFallback configures with fallback enabled or disabled
func WithFallback(fallback bool) func(*Plugin) {
	return func(p *Plugin) {
		p.fallback = fallback
	}
}

// WithFinalizeSupport adds dependencies to all pipelines to a user provider pipelined named `finalize`
func WithFinalizeSupport(finalize bool) func(*Plugin) {
	return func(p *Plugin) {
		p.finalize = finalize
	}
}

// WithAlwaysRunAll configures always run all enabled or disabled
func WithAlwaysRunAll(alwaysRunAll bool) func(*Plugin) {
	return func(p *Plugin) {
		p.alwaysRunAll = alwaysRunAll
	}
}

// WithMaxDepth configures with max depth to search for 'drone.yml'. Requires fallback to be enabled.
func WithMaxDepth(maxDepth int) func(*Plugin) {
	return func(p *Plugin) {
		p.maxDepth = maxDepth
	}
}

// WithAllowListFile configures with repo slug regex match list file
func WithAllowListFile(file string) func(*Plugin) {
	return func(p *Plugin) {
		p.allowListFile = file
	}
}

// WithConsiderFile configures with a consider file which contains references to all 'drone.yml' files which should
// be considered for the repository.
func WithConsiderFile(considerFile string) func(*Plugin) {
	return func(p *Plugin) {
		p.considerFile = considerFile
	}
}

// WithConsiderRepoConfig enable fallback to the repository 'drone.yml' when a consider file is configured but not found.
func WithConsiderRepoConfig(considerRepoConfig bool) func(*Plugin) {
	return func(p *Plugin) {
		p.considerRepoConfig = considerRepoConfig
	}
}

// WithCacheTTL enables request/response caching and the specified TTL for each entry
func WithCacheTTL(ttl time.Duration) func(*Plugin) {
	return func(p *Plugin) {
		p.cacheTTL = ttl
	}
}
