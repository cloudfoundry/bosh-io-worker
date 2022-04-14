package manifest

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	bputil "github.com/bosh-dep-forks/bosh-provisioner/util"
)

// SyntaxValidator parses and saves all manifest values to determine
// their syntactic validity. Determining if individual values make sense
// in a greater context (within a full release) is outside of scope.
type SyntaxValidator struct {
	release *Release
}

func NewSyntaxValidator(manifest *Manifest) SyntaxValidator {
	if manifest == nil {
		panic("Expected manifest to not be nil")
	}

	return SyntaxValidator{release: &manifest.Release}
}

func (v SyntaxValidator) Validate() error {
	if v.release.Name == "" {
		return bosherr.Error("Missing release name")
	}

	if v.release.Version == "" {
		return bosherr.Error("Missing release version")
	}

	if v.release.CommitHash == "" {
		return bosherr.Error("Missing release commit_hash")
	}

	for i, job := range v.release.Jobs {
		err := v.validateJob(&v.release.Jobs[i])
		if err != nil {
			return bosherr.WrapErrorf(err, "Job %s (%d)", job.Name, i)
		}
	}

	for i, pkg := range v.release.Packages {
		err := v.validatePkg(&v.release.Packages[i])
		if err != nil {
			return bosherr.WrapErrorf(err, "Package %s (%d)", pkg.Name, i)
		}
	}

	return nil
}

func (v SyntaxValidator) validateJob(job *Job) error {
	if job.Name == "" {
		return bosherr.Error("Missing name")
	}

	if job.Version == "" {
		return bosherr.Error("Missing version")
	}

	str, err := bputil.DecodePossibleBase64Str(job.Version)
	if err != nil {
		return bosherr.WrapError(err, "Decoding base64 encoded version")
	}

	job.Version = str

	if job.Fingerprint == "" {
		return bosherr.Error("Missing fingerprint")
	}

	str, err = bputil.DecodePossibleBase64Str(job.Fingerprint)
	if err != nil {
		return bosherr.WrapError(err, "Decoding base64 encoded fingerprint")
	}

	job.Fingerprint = str

	if job.SHA1 == "" {
		return bosherr.Error("Missing sha1")
	}

	str, err = bputil.DecodePossibleBase64Str(job.SHA1)
	if err != nil {
		return bosherr.WrapError(err, "Decoding base64 encoded sha1")
	}

	job.SHA1 = str

	return nil
}

func (v SyntaxValidator) validatePkg(pkg *Package) error {
	if pkg.Name == "" {
		return bosherr.Error("Missing name")
	}

	if pkg.Version == "" {
		return bosherr.Error("Missing version")
	}

	str, err := bputil.DecodePossibleBase64Str(pkg.Version)
	if err != nil {
		return bosherr.WrapError(err, "Decoding base64 encoded version")
	}

	pkg.Version = str

	if pkg.Fingerprint == "" {
		return bosherr.Error("Missing fingerprint")
	}

	str, err = bputil.DecodePossibleBase64Str(pkg.Fingerprint)
	if err != nil {
		return bosherr.WrapError(err, "Decoding base64 encoded fingerprint")
	}

	pkg.Fingerprint = str

	if pkg.SHA1 == "" {
		return bosherr.Error("Missing sha1")
	}

	str, err = bputil.DecodePossibleBase64Str(pkg.SHA1)
	if err != nil {
		return bosherr.WrapError(err, "Decoding base64 encoded sha1")
	}

	pkg.SHA1 = str

	return nil
}
