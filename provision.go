package urknall

func Provision(host Host, list *PackageList) error {
	return ProvisionWithOptions(host, list, nil)
}

type ProvisionOptions struct {
	DryRun bool
}

func ProvisionWithOptions(host Host, list *PackageList, opts *ProvisionOptions) error {
	e := list.precompileRunlists()
	if e != nil {
		return e
	}

	runner := &Runner{
		Host: host,
		User: host.User(),
	}
	e = runner.prepare()
	if e != nil {
		return e
	}
	return runner.provision(list)
}
