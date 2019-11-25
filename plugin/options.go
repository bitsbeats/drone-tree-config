package plugin

// WithAuthServer configures an auth server
func WithAuthServer(authServer string) func(*Plugin) {
	return func(p *Plugin) {
		p.authServer = authServer
	}
}

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

// WithRegexFile configures with repo slug regex match list file
func WithRegexFile(regexFile string) func(*Plugin) {
	return func(p *Plugin) {
		p.regexFile = regexFile
	}
}
