package stackgo

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Database struct {
	Name string `json:"name"`
}

func (d Database) Create() error {
	log.Println("Create database:", d.Name)

	out, err := exec.Command("sudo", "-u", "postgres", "createdb", d.Name).CombinedOutput()

	if strings.HasSuffix(strings.TrimSpace(string(out)), "already exists") {
		return nil
	}

	if err != nil {
		log.Println(string(out))
		return err
	}

	return nil
}

type Manifest struct {
	SourceLists     []SourceList             `json:"source_lists"`
	Packages        []Package                `json:"packages"`
	Templates       []Template               `json:"templates"`
	PackageArchives []PersonalPackageArchive `json:"personal_package_archives"`
	Users           []User                   `json:"users"`
	Services        []Service                `json:"services"`
	Databases       []Database               `json:"postgres_databases"`
}

func Analyze() (Manifest, error) {
	path := "/etc/apt/sources.list.d"

	_, err := ioutil.ReadDir(path)

	if err != nil {
		return Manifest{}, err
	}

	return Manifest{}, nil
}

func (m Manifest) Begin() error {
	err := os.MkdirAll("/var/stackgo", 0777)

	if err != nil {
		return err
	}

	info, err := os.Stat("/var/stackgo/apt-update")

	if os.IsNotExist(err) {
		_, err = os.Create("/var/stackgo/apt-update")

		if err != nil {
			return err
		}

		log.Println("Run `apt-get update`")
		out, err := exec.Command("apt-get", "update").CombinedOutput()

		if err != nil {
			log.Println(string(out))
			return err
		}

		return nil
	}

	// If the ModTime on this file is older than a week, rerun
	if info.ModTime().Before(time.Now().AddDate(0, 0, -7)) {

		log.Println("Run `apt-get update`")
		out, err := exec.Command("apt-get", "update").CombinedOutput()

		if err != nil {
			log.Println(string(out))
			return err
		}

	}

	return nil

}

func (m Manifest) Converge() error {
	for _, user := range m.Users {
		err := user.Create()

		if err != nil {
			return err
		}
	}

	err := m.Begin()

	if err != nil {
		return err
	}

	apt_update_needed := false

	for _, slist := range m.SourceLists {
		created, err := slist.Install()

		if err != nil {
			return err
		}

		if created {
			apt_update_needed = true
		}
	}

	// If there are Personal Package Archives to install,
	// make sure that the `add-apt-repository` command is available
	if len(m.PackageArchives) > 0 {
		pkg := Package{Name: "python-software-properties"}
		err := pkg.Install()

		if err != nil {
			return err
		}
	}

	for _, ppa := range m.PackageArchives {
		created, err := ppa.Install()

		if err != nil {
			return err
		}

		if created {
			apt_update_needed = true
		}
	}

	// Replace this with notifications eventually
	if apt_update_needed {
		log.Println("Run `apt-get update`")
		out, err := exec.Command("apt-get", "update").CombinedOutput()

		if err != nil {
			log.Println(string(out))
			return err
		}
	}

	for _, pack := range m.Packages {
		err := pack.Install()

		if err != nil {
			return err
		}
	}

	for _, tmpl := range m.Templates {
		err := tmpl.Create()

		if err != nil {
			return err
		}
	}

	for _, service := range m.Services {
		err := service.Create()

		if err != nil {
			return err
		}
	}

	for _, db := range m.Databases {
		err := db.Create()

		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Manifest) Add(other Manifest) {
	m.SourceLists = append(m.SourceLists, other.SourceLists...)
	m.Packages = append(m.Packages, other.Packages...)
	m.Templates = append(m.Templates, other.Templates...)
	m.PackageArchives = append(m.PackageArchives, other.PackageArchives...)
	m.Users = append(m.Users, other.Users...)
	m.Services = append(m.Services, other.Services...)
	m.Databases = append(m.Databases, other.Databases...)
}
