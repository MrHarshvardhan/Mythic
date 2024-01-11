package internal

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

func InstallFolder(installPath string, overWrite bool) error {
	workingPath := getCwdFromExe()
	if fileExists(filepath.Join(installPath, "config.json")) {
		var config = viper.New()
		config.SetConfigName("config")
		config.SetConfigType("json")
		fmt.Printf("[*] Parsing config.json\n")
		config.AddConfigPath(installPath)
		if err := config.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				fmt.Printf("[-] Error while reading in config file: %s", err)
				return err
			} else {
				fmt.Printf("[-] Error while parsing config file: %s", err)
				return err
			}
		}
		if !config.GetBool("exclude_payload_type") {
			// handle the payload type copying here
			files, err := ioutil.ReadDir(filepath.Join(installPath, "Payload_Type"))
			if err != nil {
				fmt.Printf("[-] Failed to list contents of new Payload_Type folder: %v\n", err)
				return err
			}
			for _, f := range files {
				if f.IsDir() {
					fmt.Printf("[*] Processing Payload Type %s\n", f.Name())
					if dirExists(filepath.Join(workingPath, InstalledServicesFolder, f.Name())) {
						if overWrite || askConfirm("[*] "+f.Name()+" already exists. Replace current version? ") {
							fmt.Printf("[*] Stopping current container\n")
							if isServiceRunning(strings.ToLower(f.Name())) {
								if err := DockerStop([]string{f.Name()}); err != nil {
									fmt.Printf("[-] Failed to stop current container: %v\n", err)
									return err
								}
							}
							fmt.Printf("[*] Removing current version\n")
							err = os.RemoveAll(filepath.Join(workingPath, InstalledServicesFolder, f.Name()))
							if err != nil {
								fmt.Printf("[-] Failed to remove current version: %v\n", err)
								fmt.Printf("[-] Continuing to the next payload\n")
								continue
							} else {
								fmt.Printf("[+] Successfully removed the current version\n")
							}
						} else {
							fmt.Printf("[!] Skipping Payload Type, %s\n", f.Name())
							continue
						}
					}
					fmt.Printf("[*] Copying new version of payload into place\n")
					err = copyDir(filepath.Join(installPath, "Payload_Type", f.Name()), filepath.Join(workingPath, InstalledServicesFolder, f.Name()))
					if err != nil {
						fmt.Printf("[-] Failed to copy directory over: %v\n", err)
						continue
					}
					fmt.Printf("[*] Adding service into docker-compose\n")
					if config.IsSet("docker-compose") {
						if err := AddDockerComposeEntry(f.Name(), config.GetStringMap("docker-compose")); err != nil {
							fmt.Printf("[-] Failed to add service to docker-compose: %v\n", err)
						} else if err := DockerBuild([]string{f.Name()}); err != nil {
							fmt.Printf("[-] Failed to start service: %v\n", err)
						}
					} else {
						if err := AddDockerComposeEntry(f.Name(), make(map[string]interface{})); err != nil {
							fmt.Printf("[-] Failed to add service to docker-compose: %v\n", err)
						} else if err := DockerBuild([]string{f.Name()}); err != nil {
							fmt.Printf("[-] Failed to start service: %v\n", err)
						}
					}

				}
			}
			fmt.Printf("[+] Successfully installed service\n")
		} else {
			fmt.Printf("[*] Skipping over Payload Type\n")
		}
		if !config.GetBool("exclude_c2_profiles") {
			// handle the c2 profile copying here
			files, err := ioutil.ReadDir(filepath.Join(installPath, "C2_Profiles"))
			if err != nil {
				fmt.Printf("[-] Failed to list contents of C2_Profiles folder from clone\n")
				return err
			}
			for _, f := range files {
				if f.IsDir() {
					fmt.Printf("[*] Processing C2 Profile %s\n", f.Name())
					if dirExists(filepath.Join(workingPath, InstalledServicesFolder, f.Name())) {
						if overWrite || askConfirm("[*] "+f.Name()+" already exists. Replace current version? ") {
							fmt.Printf("[*] Stopping current container\n")
							if isServiceRunning(strings.ToLower(f.Name())) {
								if err := DockerStop([]string{f.Name()}); err != nil {
									fmt.Printf("[-] Failed to stop container: %v\n", err)
									return err
								}
							}
							fmt.Printf("[*] Removing current version\n")
							err = os.RemoveAll(filepath.Join(workingPath, InstalledServicesFolder, f.Name()))
							if err != nil {
								fmt.Printf("[-] Failed to remove current version: %v\n", err)
								fmt.Printf("[-] Continuing to the next c2 profile\n")
								continue
							} else {
								fmt.Printf("[+] Successfully removed the current version\n")
							}
						} else {
							fmt.Printf("[!] Skipping C2 Profile, %s\n", f.Name())
							continue
						}
					}
					fmt.Printf("[*] Copying new version into place\n")
					err = copyDir(filepath.Join(installPath, "C2_Profiles", f.Name()), filepath.Join(workingPath, InstalledServicesFolder, f.Name()))
					if err != nil {
						fmt.Printf("[-] Failed to copy directory over\n")
						continue
					}
					// now add payload type to yaml config
					fmt.Printf("[*] Adding c2, %s, into docker-compose\n", f.Name())
					if err := AddDockerComposeEntry(f.Name(), make(map[string]interface{})); err != nil {
						fmt.Printf("[-] Failed to add %s to docker-compose: %v\n", f.Name(), err)
					} else if err := DockerBuild([]string{f.Name()}); err != nil {
						fmt.Printf("[-] Failed to start service: %v\n", err)
					}
				}
			}
			fmt.Printf("[+] Successfully installed c2\n")
		} else {
			fmt.Printf("[*] Skipping over C2 Profile\n")
		}
		if !config.GetBool("exclude_documentation_payload") {
			// handle payload documentation copying here
			files, err := ioutil.ReadDir(filepath.Join(installPath, "documentation-payload"))
			if err != nil {
				fmt.Printf("[-] Failed to list contents of documentation_payload folder from clone: %v\n", err)
			} else {
				for _, f := range files {
					if f.IsDir() {
						fmt.Printf("[*] Processing Documentation for %s\n", f.Name())
						if mythicEnv.GetBool("documentation_bind_local_mount") {
							if dirExists(filepath.Join(workingPath, "documentation-docker", "content", "Agents", f.Name())) {
								if overWrite || askConfirm("[*] "+f.Name()+" documentation already exists. Replace current version? ") {
									fmt.Printf("[*] Removing current version\n")
									err = os.RemoveAll(filepath.Join(workingPath, "documentation-docker", "content", "Agents", f.Name()))
									if err != nil {
										fmt.Printf("[-] Failed to remove current version: %v\n", err)
										fmt.Printf("[-] Continuing to the next payload documentation\n")
										continue
									} else {
										fmt.Printf("[+] Successfully removed the current version\n")
									}
								} else {
									fmt.Printf("[!] Skipping documentation for , %s\n", f.Name())
									continue
								}
							}
							fmt.Printf("[*] Copying new documentation into place\n")
							err = copyDir(filepath.Join(installPath, "documentation-payload", f.Name()), filepath.Join(workingPath, "documentation-docker", "content", "Agents", f.Name()))
							if err != nil {
								fmt.Printf("[-] Failed to copy directory over\n")
								continue
							}
						} else {
							err = moveFileToVolume("mythic_documentation_volume",
								filepath.Join("content", "Agents"),
								filepath.Join(installPath, "documentation-payload", f.Name()))
							if err != nil {
								fmt.Printf("[-] Failed to install documentation for payload: %v\n", err)
								continue
							}
						}

					}
				}
				fmt.Printf("[+] Successfully installed Payload documentation\n")
			}

		} else {
			fmt.Printf("[*] Skipping over Payload Documentation\n")
		}
		if !config.GetBool("exclude_documentation_c2") {
			// handle the c2 documentation copying here
			files, err := ioutil.ReadDir(filepath.Join(installPath, "documentation-c2"))
			if err != nil {
				fmt.Printf("[-] Failed to list contents of documentation_payload folder from clone")
			} else {
				for _, f := range files {
					if f.IsDir() {
						fmt.Printf("[*] Processing Documentation for %s\n", f.Name())
						if mythicEnv.GetBool("document_bind_local_mount") {
							if dirExists(filepath.Join(workingPath, "documentation-docker", "content", "C2 Profiles", f.Name())) {
								if overWrite || askConfirm("[*] "+f.Name()+" documentation already exists. Replace current version? ") {
									fmt.Printf("[*] Removing current version\n")
									err = os.RemoveAll(filepath.Join(workingPath, "documentation-docker", "content", "C2 Profiles", f.Name()))
									if err != nil {
										fmt.Printf("[-] Failed to remove current version: %v\n", err)
										fmt.Printf("[-] Continuing to the next c2 documentation\n")
										continue
									} else {
										fmt.Printf("[+] Successfully removed the current version\n")
									}
								} else {
									fmt.Printf("[!] Skipping documentation for %s\n", f.Name())
									continue
								}
							}
							fmt.Printf("[*] Copying new documentation version into place\n")
							err = copyDir(filepath.Join(installPath, "documentation-c2", f.Name()), filepath.Join(workingPath, "documentation-docker", "content", "C2 Profiles", f.Name()))
							if err != nil {
								fmt.Printf("[-] Failed to copy directory over\n")
								continue
							}
						} else {
							err = moveFileToVolume("mythic_documentation_volume",
								filepath.Join("content", "C2 Profiles"),
								filepath.Join(installPath, "documentation-c2", f.Name()))
							if err != nil {
								fmt.Printf("[-] Failed to install documentation for c2: %v\n", err)
								continue
							}

						}

					}
				}
				fmt.Printf("[+] Successfully installed c2 documentation\n")
			}

		} else {
			fmt.Printf("[*] Skipping over C2 Documentation\n")
		}
		if !config.GetBool("exclude_documentation_wrapper") {
			// handle payload documentation copying here
			files, err := ioutil.ReadDir(filepath.Join(installPath, "documentation-wrapper"))
			if err != nil {
				fmt.Printf("[-] Failed to list contents of documentation-wrapper folder from clone: %v\n", err)
			} else {
				for _, f := range files {
					if f.IsDir() {
						fmt.Printf("[*] Processing Documentation for %s\n", f.Name())
						if mythicEnv.GetBool("document_local_bind_mount") {
							if dirExists(filepath.Join(workingPath, "documentation-docker", "content", "Wrappers", f.Name())) {
								if overWrite || askConfirm("[*] "+f.Name()+" documentation already exists. Replace current version? ") {
									fmt.Printf("[*] Removing current version\n")
									err = os.RemoveAll(filepath.Join(workingPath, "documentation-docker", "content", "Wrappers", f.Name()))
									if err != nil {
										fmt.Printf("[-] Failed to remove current version: %v\n", err)
										fmt.Printf("[-] Continuing to the next wrapper documentation\n")
										continue
									} else {
										fmt.Printf("[+] Successfully removed the current version\n")
									}
								} else {
									fmt.Printf("[!] Skipping documentation for , %s\n", f.Name())
									continue
								}
							}
							fmt.Printf("[*] Copying new documentation into place\n")
							err = copyDir(filepath.Join(installPath, "documentation-wrapper", f.Name()), filepath.Join(workingPath, "documentation-docker", "content", "Wrappers", f.Name()))
							if err != nil {
								fmt.Printf("[-] Failed to copy directory over\n")
								continue
							}
						} else {
							err = moveFileToVolume("mythic_documentation_volume",
								filepath.Join("content", "Wrappers"),
								filepath.Join(installPath, "documentation-wrapper", f.Name()))
							if err != nil {
								fmt.Printf("[-] Failed to install documentation for wrapper: %v\n", err)
								continue
							}
						}
					}
				}
				fmt.Printf("[+] Successfully installed Wrapper documentation\n")
			}
		} else {
			fmt.Printf("[*] Skipping over Wrapper Documentation\n")
		}
		if isServiceRunning("mythic_documentation") {
			fmt.Printf("[*] Restarting mythic_documentation container to pull in changes\n")
			DockerStop([]string{"mythic_documentation"})
			DockerStart([]string{"mythic_documentation"})
		}
	} else {
		log.Fatal("[-] Failed to find config.json in cloned down repo\n")
	}
	return nil
}
func InstallService(url string, branch string, overWrite bool) error {
	// make our temp directory to clone into
	workingPath := getCwdFromExe()
	fmt.Printf("[*] Creating temporary directory\n")
	if dirExists(filepath.Join(workingPath, "tmp")) {
		err := os.RemoveAll(filepath.Join(workingPath, "tmp"))
		if err != nil {
			fmt.Printf("[-] tmp directory couldn't be deleted for a fresh install: %v\n", err)
			return err
		}
	}
	err := os.Mkdir(filepath.Join(workingPath, "tmp"), 0755)
	defer os.RemoveAll(filepath.Join(workingPath, "tmp"))
	if err != nil {
		fmt.Printf("[-] Failed to make temp directory for cloning: %v\n", err)
		return err
	}
	if branch == "" {
		fmt.Printf("[*] Cloning %s\n", url)
		err = runGitClone([]string{"-c", "http.sslVerify=false", "clone", "--recurse-submodules", "--single-branch", url, filepath.Join(workingPath, "tmp")})
	} else {
		fmt.Printf("[*] Cloning branch \"%s\" from %s\n", branch, url)
		err = runGitClone([]string{"-c", "http.sslVerify=false", "clone", "--recurse-submodules", "--single-branch", "--branch", branch, url, filepath.Join(workingPath, "tmp")})
	}
	if err != nil {
		fmt.Printf("[-] Failed to clone down repository: %v\n", err)
		return err
	}
	if err = InstallFolder(filepath.Join(workingPath, "tmp"), overWrite); err != nil {
		fmt.Printf("[-] Failed to install: %v\n", err)
		return err
	} else {
		return nil
	}
}
func installServiceByName(service string) error {
	// just have a service name, check MythicAgents, MythicC2Profiles, or error
	agentURL := fmt.Sprintf("https://github.com/MythicAgents/%s", service)
	c2URL := fmt.Sprintf("https://github.com/MythicC2Profiles/%s", service)
	if err := runGitLsRemote([]string{"lfs-remote", agentURL, "HEAD"}); err != nil {
		if err := runGitLsRemote([]string{"lfs-remote", c2URL, "HEAD"}); err != nil {
			fmt.Printf("[-] Failed to find an agent or c2 profile by that name")
			return err
		} else {
			// this exists as a c2 profile repo, so we can pull that down
			return InstallService(c2URL, "", true)
		}
	} else {
		// this exists as an agent repo, so we can pull that down
		return InstallService(agentURL, "", true)
	}
}
func InstallMythicSyncFolder(installPath string) error {
	workingPath := getCwdFromExe()
	viper.SetConfigName("docker-compose")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(getCwdFromExe())
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Printf("[-] Error while reading in docker-compose file: %s\n", err)
			return err
		} else {
			fmt.Printf("[-] Error while parsing docker-compose file: %s\n", err)
			return err
		}
	}
	service := "mythic_sync"
	if isServiceRunning(service) {
		if err := DockerStop([]string{service}); err != nil {
			fmt.Printf("[-] Failed to stop current docker container: %v\n", err)
			return err
		}
	}
	if dirExists(filepath.Join(workingPath, InstalledServicesFolder, service)) {
		err := os.RemoveAll(filepath.Join(workingPath, InstalledServicesFolder, service))
		if err != nil {
			fmt.Printf("[-] %s directory couldn't be deleted for a fresh install: %v\n", filepath.Join(workingPath, InstalledServicesFolder, service), err)
			return err
		}
	}
	if err := copyDir(installPath, filepath.Join(workingPath, InstalledServicesFolder, service)); err != nil {
		fmt.Printf("[-] Failed to create %s directory to install mythic_sync: %v\n", service, err)
		return err
	}
	addMythicServiceDockerComposeEntry(service)
	fmt.Printf("[+] Successfully installed mythic_sync!\n")
	if isServiceRunning("mythic_server") {
		if err := DockerStart([]string{strings.ToLower(service)}); err != nil {
			fmt.Printf("[-] Failed to start mythic_sync: %v\n", err)
		}
	}
	return nil
}
func InstallMythicSync(url string, branch string) error {
	// make our temp directory to clone into
	workingPath := getCwdFromExe()
	fmt.Printf("[*] Creating temporary directory\n")
	if dirExists(filepath.Join(workingPath, "tmp")) {
		if err := os.RemoveAll(filepath.Join(workingPath, "tmp")); err != nil {
			fmt.Printf("[-] %s directory couldn't be deleted for a fresh install: %v\n", filepath.Join(workingPath, "tmp"), err)
			return err
		}
	}
	err := os.Mkdir(filepath.Join(workingPath, "tmp"), 0755)
	defer os.RemoveAll(filepath.Join(workingPath, "tmp"))
	if err != nil {
		log.Fatalf("[-] Failed to make temp directory for cloning: %v\n", err)
	}
	if branch == "" {
		fmt.Printf("[*] Cloning %s\n", url)
		err = runGitClone([]string{"-c", "http.sslVerify=false", "clone", "--recurse-submodules", "--single-branch", url, filepath.Join(workingPath, "tmp")})
	} else {
		fmt.Printf("[*] Cloning branch \"%s\" from %s\n", branch, url)
		err = runGitClone([]string{"-c", "http.sslVerify=false", "clone", "--recurse-submodules", "--single-branch", "--branch", branch, url, filepath.Join(workingPath, "tmp")})
	}
	if err != nil {
		fmt.Printf("[-] Failed to clone down repository: %v\n", err)
		return err
	}
	return InstallMythicSyncFolder(filepath.Join(workingPath, "tmp"))

}
func UninstallMythicSync() {
	workingPath := getCwdFromExe()
	service := "mythic_sync"
	if isServiceRunning(service) {
		DockerStop([]string{service})
	}
	removeMythicServiceDockerComposeEntry(service)
	if dirExists(filepath.Join(workingPath, InstalledServicesFolder, service)) {
		err := os.RemoveAll(filepath.Join(workingPath, InstalledServicesFolder, service))
		if err != nil {
			log.Fatalf("[-] %s directory couldn't be deleted: %v\n", service, err)
		} else {
			fmt.Printf("[+] Successfully removed %s from disk\n", service)
		}
	} else {
		fmt.Printf("[+] %s was not installed on disk\n", service)
	}
	fmt.Printf("[+] Successfully uninstalled mythic_sync\n")
}
func UninstallService(services []string) {
	workingPath := getCwdFromExe()
	for _, service := range services {
		if stringInSlice(strings.ToLower(service), MythicPossibleServices) {
			fmt.Printf("[-] Trying to uninstall Mythic services not allowed\n")
			os.Exit(1)
		}
		found := false
		if dirExists(filepath.Join(workingPath, InstalledServicesFolder, service)) {
			fmt.Printf("[*] Stopping and removing container\n")
			if isServiceRunning(strings.ToLower(service)) {
				if err := DockerStop([]string{strings.ToLower(service)}); err != nil {
					fmt.Printf("[-] Failed to stop container: %v\n", err)
					return
				}
			}
			fmt.Printf("[*] Removing %s from docker-compose\n", strings.ToLower(service))
			if err := RemoveDockerComposeEntry(strings.ToLower(service)); err != nil {
				fmt.Printf("[-] Failed to remove docker compose entry: %v\n", err)
				return
			}
			fmt.Printf("[*] Removing Payload Type folder from disk\n")
			found = true
			err := os.RemoveAll(filepath.Join(workingPath, InstalledServicesFolder, service))
			if err != nil {
				fmt.Printf("[-] Failed to remove folder: %v\n", err)
				os.Exit(1)
			} else {
				fmt.Printf("[+] Successfully removed %s's folder\n", service)
			}
			if dirExists(filepath.Join(workingPath, "documentation-docker", "content", "Agents", service)) {
				fmt.Printf("[*] Removing Payload Type's Documentation from disk\n")
				err = os.RemoveAll(filepath.Join(workingPath, "documentation-docker", "content", "Agents", service))
				if err != nil {
					fmt.Printf("[-] Failed to remove Payload Type's Documentation: %v\n", err)
					os.Exit(1)
				} else {
					fmt.Printf("[+] Successfully removed Payload Type's Documentation\n")
				}
			}
			if dirExists(filepath.Join(workingPath, "documentation-docker", "content", "C2 Profiles", service)) {
				fmt.Printf("[*] Removing C2 Profile's Documentation\n")
				err = os.RemoveAll(filepath.Join(workingPath, "documentation-docker", "content", "C2 Profiles", service))
				if err != nil {
					fmt.Printf("[-] Failed to remove C2 Profile's Documentation: %v\n", err)
					os.Exit(1)
				} else {
					fmt.Printf("[+] Successfully removed C2 Profile's Documentation\n")
				}
			}
			if dirExists(filepath.Join(workingPath, "documentation-docker", "content", "Wrappers", service)) {
				fmt.Printf("[*] Removing C2 Profile's Documentation\n")
				err = os.RemoveAll(filepath.Join(workingPath, "documentation-docker", "content", "Wrappers", service))
				if err != nil {
					fmt.Printf("[-] Failed to remove C2 Profile's Documentation: %v\n", err)
					os.Exit(1)
				} else {
					fmt.Printf("[+] Successfully removed C2 Profile's Documentation\n")
				}
			}
			if fileExists(filepath.Join(workingPath, "mythic-docker", "src", "static", service+".svg")) {
				found = true
				err := os.RemoveAll(filepath.Join(workingPath, "src", "static", service+".svg"))
				if err != nil {
					fmt.Printf("[-] Failed to agent icon: %v\n", err)
					os.Exit(1)
				} else {
					fmt.Printf("[+] Successfully removed %s's old UI icon\n", service)
				}
			}
		}

		if found {
			fmt.Printf("[+] Successfully Uninstalled %s\n", service)
			if isServiceRunning("mythic_documentation") {
				fmt.Printf("[*] Restarting mythic_documentation container to pull in changes\n")
				DockerStop([]string{"mythic_documentation"})
				DockerStart([]string{"mythic_documentation"})
			}
			return
		} else {
			fmt.Printf("[-] Failed to find any service folder by that name\n")
			os.Exit(1)
		}
	}
}
