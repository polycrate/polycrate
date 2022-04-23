package cmd

// func (c ChartConfig) Download(fp string) error {
// 	// If a URL is set for this chart, download it directly
// 	if c.Url != "" {
// 		log.Debug("Found URL: " + c.Url)
// 		err := DownloadFile(c.Url, fp)
// 		return err
// 		// If no URL is set, determine to correct chart URL from the repository, if given
// 	} else if c.Repo.Url != "" {
// 		// Create TempFile for Chart Repository index.yaml
// 		chartRepoIndexFile, err := getTempFile("yaml")
// 		CheckErr(err)

// 		// Download Chart Repository index.yaml to Tempfile
// 		err = DownloadFile(c.Repo.Url+"/index.yaml", chartRepoIndexFile.Name())
// 		CheckErr(err)

// 		// Create new config instance for the repository index
// 		chartRepoIndex, err := loadConfigFile(chartRepoIndexFile.Name(), "yaml")
// 		CheckErr(err)

// 		// Get the list of available charts for this plugin
// 		pluginCharts := chartRepoIndex.Get("entries." + c.Name)

// 		// pluginCharts contains a slice but is of type interface{}
// 		// as such "range" won't work here and we need to fall back
// 		// to plain old for loops to get the config for each chart
// 		//
// 		// Example:
// 		//
// 		// - annotations:
// 		// 		category: Infrastructure
// 		// 	apiVersion: v2
// 		// 	appVersion: 2.4.1
// 		// 	created: "2021-10-28T05:44:20.489968646Z"
// 		// 	dependencies:
// 		// 	- name: common
// 		// 		repository: https://charts.bitnami.com/bitnami
// 		// 		tags:
// 		// 		- bitnami-common
// 		// 		version: 1.x.x
// 		// 	- name: postgresql
// 		// 		repository: https://charts.bitnami.com/bitnami
// 		// 		version: 10.x.x
// 		// 	- condition: redis.enabled
// 		// 		name: redis
// 		// 		repository: https://charts.bitnami.com/bitnami
// 		// 		version: 15.x.x
// 		// 	description: Kubeapps is a dashboard for your Kubernetes cluster that makes it
// 		// 		easy to deploy and manage applications in your cluster using Helm
// 		// 	digest: aa442fc2924186944f44f38ce301b445556c2d6f951cef1472657e51b3df6c25
// 		// 	home: https://kubeapps.com
// 		// 	icon: https://raw.githubusercontent.com/kubeapps/kubeapps/master/docs/img/logo.png
// 		// 	keywords:
// 		// 	- helm
// 		// 	- dashboard
// 		// 	- service catalog
// 		// 	- deployment
// 		// 	maintainers:
// 		// 	- email: containers@bitnami.com
// 		// 		name: Bitnami
// 		// 	name: kubeapps
// 		// 	sources:
// 		// 	- https://github.com/kubeapps/kubeapps
// 		// 	urls:
// 		// 	- https://charts.bitnami.com/bitnami/kubeapps-7.5.10.tgz
// 		// 	version: 7.5.10
// 		switch reflect.TypeOf(pluginCharts).Kind() {
// 		case reflect.Slice:
// 			s := reflect.ValueOf(pluginCharts)

// 			for i := 0; i < s.Len(); i++ {
// 				// pluginChart is a map containing the values shown above
// 				pluginChart := s.Index(i).Interface()
// 				pc := pluginChart.(map[interface{}]interface{})

// 				// Obtain the chart urls
// 				urls := pc["urls"]

// 				switch reflect.TypeOf(urls).Kind() {
// 				case reflect.Slice:
// 					s := reflect.ValueOf(urls)
// 					url := s.Index(0).String()
// 					fmt.Println(url)

// 					// Download the chart
// 					err := DownloadFile(url, fp)
// 					return err
// 				}
// 			}
// 		}
// 		return nil
// 	}
// 	return errors.New("unknown error occured")
// }
