package cmd

type PolycrateGitlabProvider struct {
	Transport string `yaml:"transport,omitempty" mapstructure:"transport,omitempty" json:"transport,omitempty"`
	Url       string `yaml:"url,omitempty" mapstructure:"url,omitempty" json:"url,omitempty"`
	Group     string `yaml:"group,omitempty" mapstructure:"group,omitempty" json:"group,omitempty"`
	Token     string `yaml:"token,omitempty" mapstructure:"token,omitempty" json:"token,omitempty"`
	// environment variables to check for gitlab token
	TokenEnv []string `yaml:"token_env,omitempty" mapstructure:"token_env,omitempty" json:"token_env,omitempty"`
}

// func (p PolycrateGitlabProvider) Print() {
// 	printObject(p)
// }

// func (p PolycrateGitlabProvider) GetCredentials() (PolycrateProviderCredentials, error) {
// 	log.WithFields(log.Fields{
// 		"provider": "gitlab",
// 	}).Debugf("Getting credentials")
// 	credentials := PolycrateProviderCredentials{
// 		username: "",
// 		password: "",
// 	}

// 	// Check for token
// 	if p.Token != "" {
// 		log.WithFields(log.Fields{
// 			"provider": "gitlab",
// 		}).Debugf("Using token from config")
// 		credentials.password = p.Token
// 		return credentials, nil
// 	}

// 	// Check for token in environment variables
// 	for _, envVar := range p.TokenEnv {
// 		token := os.Getenv(envVar)

// 		if token != "" {
// 			credentials.password = token
// 			log.WithFields(log.Fields{
// 				"provider": "gitlab",
// 			}).Debugf("Using token from env var %s", envVar)
// 			return credentials, nil
// 		}
// 	}

// 	err := errors.New("no credentials found")

// 	return credentials, err
// }

// func (p PolycrateGitlabProvider) GetDefaultGroup() (PolycrateProviderGroup, error) {
// 	group, err := p.GetGroup(p.Group)
// 	if err != nil {
// 		return PolycrateProviderGroup{}, err
// 	}
// 	return group, nil
// }

// func (p PolycrateGitlabProvider) GetName() string {
// 	return "gitlab"
// }

// func (p PolycrateGitlabProvider) GetGroups() ([]PolycrateProviderGroup, error) {
// 	log.WithFields(log.Fields{
// 		"provider": "gitlab",
// 	}).Debugf("Loading groups")
// 	git, err := p.GetClient()
// 	if err != nil {
// 		return nil, err
// 	}

// 	groups, _, err := git.Groups.ListGroups(&gitlab.ListGroupsOptions{})
// 	if err != nil {
// 		return nil, err
// 	}

// 	sort.Slice(groups, func(i, j int) bool {
// 		// If a namespace is set, push this to the top of the list
// 		if groups[i].FullPath == config.Gitlab.Group {
// 			return true
// 		}
// 		// otherwise, sort alphabetically
// 		return groups[i].FullName < groups[j].FullName
// 	})

// 	results := []PolycrateProviderGroup{}

// 	for _, group := range groups {
// 		g := PolycrateProviderGroup{
// 			name: group.FullName,
// 			url:  group.WebURL,
// 			id:   group.ID,
// 			path: group.FullPath,
// 		}
// 		results = append(results, g)
// 	}

// 	log.WithFields(log.Fields{
// 		"provider": "gitlab",
// 	}).Debugf("Groups loaded")

// 	return results, nil
// }

// func (p PolycrateGitlabProvider) GetGroup(name string) (PolycrateProviderGroup, error) {
// 	log.WithFields(log.Fields{
// 		"provider": "gitlab",
// 		"group":    name,
// 	}).Debugf("Loading group")

// 	git, err := p.GetClient()
// 	if err != nil {
// 		return PolycrateProviderGroup{}, err
// 	}

// 	groups, _, err := git.Groups.SearchGroup(name)
// 	if err != nil {
// 		return PolycrateProviderGroup{}, err
// 	}

// 	if len(groups) <= 0 {
// 		return PolycrateProviderGroup{}, fmt.Errorf("no group with name %s found", name)
// 	}

// 	log.WithFields(log.Fields{
// 		"provider": "gitlab",
// 		"group":    name,
// 	}).Debugf("Group loaded")

// 	group := PolycrateProviderGroup{
// 		name: name,
// 		url:  "",
// 		id:   groups[0].ID,
// 		path: groups[0].FullPath,
// 	}

// 	return group, nil
// }

// func (p PolycrateGitlabProvider) GetRepository(name string) (PolycrateProviderProject, error) {
// 	log.WithFields(log.Fields{
// 		"provider":   "gitlab",
// 		"repository": name,
// 	}).Debugf("Loading repository from provider")

// 	git, err := p.GetClient()
// 	if err != nil {
// 		return PolycrateProviderProject{}, err
// 	}

// 	projects, _, err := git.Search.Projects(name, &gitlab.SearchOptions{})
// 	if err != nil {
// 		return PolycrateProviderProject{}, err
// 	}

// 	if len(projects) <= 0 {
// 		return PolycrateProviderProject{}, fmt.Errorf("project not found: %s", name)
// 	}

// 	log.WithFields(log.Fields{
// 		"provider": "gitlab",
// 		"name":     name,
// 	}).Debugf("Group loaded")
// 	printObject(projects[0])

// 	project := PolycrateProviderProject{
// 		name: name,
// 		url:  "",
// 		id:   projects[0].ID,
// 		path: projects[0].PathWithNamespace,
// 	}

// 	return project, nil
// }

// func (p PolycrateGitlabProvider) CreateProject(group PolycrateProviderGroup, name string) (PolycrateProviderProject, error) {
// 	log.WithFields(log.Fields{
// 		"provider": "gitlab",
// 		"group":    group.path,
// 		"name":     name,
// 	}).Debugf("Creating project")
// 	git, err := p.GetClient()
// 	if err != nil {
// 		return PolycrateProviderProject{}, err
// 	}

// 	options := gitlab.CreateProjectOptions{
// 		Name:        &name,
// 		NamespaceID: &group.id,
// 	}

// 	project, _, err := git.Projects.CreateProject(&options)
// 	printObject(project)
// 	if err != nil {
// 		return PolycrateProviderProject{}, err
// 	}

// 	result := PolycrateProviderProject{
// 		name:        name,
// 		url:         project.WebURL,
// 		remote_ssh:  project.SSHURLToRepo,
// 		remote_http: project.HTTPURLToRepo,
// 		path:        project.PathWithNamespace,
// 		id:          project.ID,
// 	}

// 	log.WithFields(log.Fields{
// 		"provider": "gitlab",
// 		"group":    group.path,
// 		"name":     name,
// 		"url":      result.url,
// 	}).Debugf("Created project")

// 	return result, nil
// }

// func (p *PolycrateGitlabProvider) GetClient() (*gitlab.Client, error) {
// 	log.WithFields(log.Fields{
// 		"provider": "gitlab",
// 	}).Debugf("Loading API client")
// 	var git *gitlab.Client
// 	var url string

// 	credentials, err := p.GetCredentials()

// 	if err != nil {
// 		return nil, err
// 	}

// 	// Update base url
// 	if config.Gitlab.Url != "" {
// 		url = config.Gitlab.Url
// 	} else {
// 		url = "https://gitlab.com"
// 	}

// 	git, err = gitlab.NewClient(credentials.password, gitlab.WithBaseURL(config.Gitlab.Url))
// 	if err != nil {
// 		return nil, err
// 	}
// 	log.WithFields(log.Fields{
// 		"provider": "gitlab",
// 		"url":      url,
// 	}).Debugf("API client loaded")

// 	return git, nil
// }
