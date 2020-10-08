package plugin

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

// WithMaxDepth configures with max depth to search for 'drone.yml'. Requires fallback to be enabled.
func WithMaxDepth(maxDepth int) func(*Plugin) {
	return func(p *Plugin) {
		p.maxDepth = maxDepth
	}
}

// WithWhitelistFile configures with repo slug regex match list file
// Deprecated: Use WithAllowlistFile instead.
func WithWhitelistFile(file string) func(*Plugin) {
	return WithAllowListFile(file)
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
