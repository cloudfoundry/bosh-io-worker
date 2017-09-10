package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		panic(fmt.Sprintf("Wrong args: bosh-io-releases-index-dir(eg releases-index)"))
	}

	err := ReleaseVersions{ReleasesIndexDir: os.Args[1]}.Import(os.Stdin)
	if err != nil {
		panic(fmt.Sprintf("Failed: %s", err))
	}
}

type ReleaseVersions struct {
	ReleasesIndexDir string
}

type key struct {
	Source     string
	VersionRaw string
}

type value map[string]interface{}

// eg  {"Source":"github.com/cloudfoundry/bosh","VersionRaw":"138"} | {"Name":"bosh","Version":"138","CommitHash":"84561d56","UncommittedChanges":true,"Jobs":[{"Name":"health_monitor","Version":"63aad95fa56011b531ae915f7de4f562645453d4","Fingerprint":"63aad95fa56011b531ae915f7de4f562645453d4","SHA1":"eacdb519d6edf90481289998f327dd03af96f9f6","TarPath":"/mnt/tmp/tar-CmdExtractor550036964/jobs/health_monitor.tgz"},{"Name":"registry","Version":"fedffa77d6bcc67a6ff0afc0014f6b4006ce86a8","Fingerprint":"fedffa77d6bcc67a6ff0afc0014f6b4006ce86a8","SHA1":"5ccd09a20e0277ab443920da75212e0d1f4c1a0f","TarPath":"/mnt/tmp/tar-CmdExtractor550036964/jobs/registry.tgz"},{"Name":"postgres","Version":"90e43294de79754e659fc1731f515dc090f922f0","Fingerprint":"90e43294de79754e659fc1731f515dc090f922f0","SHA1":"cc7c6e1f2427339700641883a46dab8d50bf3429","TarPath":"/mnt/tmp/tar-CmdExtractor550036964/jobs/postgres.tgz"},{"Name":"blobstore","Version":"ec529e2dceecbf1e39679a3a53bbdccfcac0fd44","Fingerprint":"ec529e2dceecbf1e39679a3a53bbdccfcac0fd44","SHA1":"9bc537b7fe9713131117bcd24f29219df6d922e7","TarPath":"/mnt/tmp/tar-CmdExtractor550036964/jobs/blobstore.tgz"},{"Name":"director","Version":"93aa5c2fd4b7a0cae4c3db8751497b1b27097a9d","Fingerprint":"93aa5c2fd4b7a0cae4c3db8751497b1b27097a9d","SHA1":"815cde9cd22882fa678d44714d322cb5f3ab5727","TarPath":"/mnt/tmp/tar-CmdExtractor550036964/jobs/director.tgz"},{"Name":"redis","Version":"8b447660541a0658b2cbed6f19932c769f037188","Fingerprint":"8b447660541a0658b2cbed6f19932c769f037188","SHA1":"8738134e8a3352dc1408384bb1883571fa3633f3","TarPath":"/mnt/tmp/tar-CmdExtractor550036964/jobs/redis.tgz"},{"Name":"powerdns","Version":"d160e92adf53a224df94286107b7bff0495f9fcd","Fingerprint":"d160e92adf53a224df94286107b7bff0495f9fcd","SHA1":"02a9d75bb3dbf306df0c8a4a5f09340b0633ca92","TarPath":"/mnt/tmp/tar-CmdExtractor550036964/jobs/powerdns.tgz"},{"Name":"nats","Version":"51186524ed10bf7110a8fd72da036e1acaf7ba17","Fingerprint":"51186524ed10bf7110a8fd72da036e1acaf7ba17","SHA1":"038787742a2d28b60a9a95f9faf9688c5b452e30","TarPath":"/mnt/tmp/tar-CmdExtractor550036964/jobs/nats.tgz"}],"Packages":[{"Name":"health_monitor","Version":"7646560ef68ce72cf7236dec173b372fb25037f2","Fingerprint":"7646560ef68ce72cf7236dec173b372fb25037f2","SHA1":"547bc761b325207996948eb375838eb96c724250","TarPath":"/mnt/tmp/tar-CmdExtractor550036964/packages/health_monitor.tgz","Dependencies":[{"Name":"ruby","Version":"8c1c0bba2f15f89e3129213e3877dd40e339592f","Fingerprint":"8c1c0bba2f15f89e3129213e3877dd40e339592f","SHA1":"2f894c461afb1586dce78818e3c252f8594736cf","TarPath":"/mnt/tmp/tar-CmdExtractor550036964/packages/ruby.tgz","Dependencies":null}]},{"Name":"nginx","Version":"8f131f14088764682ebd9ff399707f8adb9a5038","Fingerprint":"8f131f14088764682ebd9ff399707f8adb9a5038","SHA1":"99df5daf35b254992e101077bf8eb12671049c62","TarPath":"/mnt/tmp/tar-CmdExtractor550036964/packages/nginx.tgz","Dependencies":null},{"Name":"genisoimage","Version":"008d332ba1471bccf9d9aeb64c258fdd4bf76201","Fingerprint":"008d332ba1471bccf9d9aeb64c258fdd4bf76201","SHA1":"2acf8da0250db31e762a23756fd0e12cfa947eb9","TarPath":"/mnt/tmp/tar-CmdExtractor550036964/packages/genisoimage.tgz","Dependencies":null},{"Name":"registry","Version":"cdbe2cca8ec99aa33ccd1d8c5c26c7ca494e0afa","Fingerprint":"cdbe2cca8ec99aa33ccd1d8c5c26c7ca494e0afa","SHA1":"029099b206c912c1bf57bd1e530a296334d48142","TarPath":"/mnt/tmp/tar-CmdExtractor550036964/packages/registry.tgz","Dependencies":[{"Name":"libpq","Version":"6aa19afb153dc276924693dc724760664ce61593","Fingerprint":"6aa19afb153dc276924693dc724760664ce61593","SHA1":"7f0f4183dea968373f73091d1c911d5edebc3f25","TarPath":"/mnt/tmp/tar-CmdExtractor550036964/packages/libpq.tgz","Dependencies":null},{"Name":"mysql","Version":"e5309aed88f5cc662bc77988a31874461f7c4fb8","Fingerprint":"e5309aed88f5cc662bc77988a31874461f7c4fb8","SHA1":"cc2cb926b6a644d2692700c5ae3bd4f5b84dcb82","TarPath":"/mnt/tmp/tar-CmdExtractor550036964/packages/mysql.tgz","Dependencies":null},{"Name":"ruby","Version":"8c1c0bba2f15f89e3129213e3877dd40e339592f","Fingerprint":"8c1c0bba2f15f89e3129213e3877dd40e339592f","SHA1":"2f894c461afb1586dce78818e3c252f8594736cf","TarPath":"/mnt/tmp/tar-CmdExtractor550036964/packages/ruby.tgz","Dependencies":null}]},{"Name":"ruby","Version":"8c1c0bba2f15f89e3129213e3877dd40e339592f","Fingerprint":"8c1c0bba2f15f89e3129213e3877dd40e339592f","SHA1":"2f894c461afb1586dce78818e3c252f8594736cf","TarPath":"/mnt/tmp/tar-CmdExtractor550036964/packages/ruby.tgz","Dependencies":null},{"Name":"postgres","Version":"aa7f5b110e8b368eeb8f5dd032e1cab66d8614ce","Fingerprint":"aa7f5b110e8b368eeb8f5dd032e1cab66d8614ce","SHA1":"a85f2a1efa7dff34ab772d67191d23363959276a","TarPath":"/mnt/tmp/tar-CmdExtractor550036964/packages/postgres.tgz","Dependencies":null},{"Name":"blobstore","Version":"a7747eb9def2054899c8553bd1b0310b9797bb6e","Fingerprint":"a7747eb9def2054899c8553bd1b0310b9797bb6e","SHA1":"ebbe2784a181a1b7e0a633aee5f657047799daae","TarPath":"/mnt/tmp/tar-CmdExtractor550036964/packages/blobstore.tgz","Dependencies":[{"Name":"ruby","Version":"8c1c0bba2f15f89e3129213e3877dd40e339592f","Fingerprint":"8c1c0bba2f15f89e3129213e3877dd40e339592f","SHA1":"2f894c461afb1586dce78818e3c252f8594736cf","TarPath":"/mnt/tmp/tar-CmdExtractor550036964/packages/ruby.tgz","Dependencies":null}]},{"Name":"director","Version":"a504833c238d24ccb782b8c0e7f021f34a409957","Fingerprint":"a504833c238d24ccb782b8c0e7f021f34a409957","SHA1":"0957005130046e65ef4f3da7227b36fd39eb9535","TarPath":"/mnt/tmp/tar-CmdExtractor550036964/packages/director.tgz","Dependencies":[{"Name":"libpq","Version":"6aa19afb153dc276924693dc724760664ce61593","Fingerprint":"6aa19afb153dc276924693dc724760664ce61593","SHA1":"7f0f4183dea968373f73091d1c911d5edebc3f25","TarPath":"/mnt/tmp/tar-CmdExtractor550036964/packages/libpq.tgz","Dependencies":null},{"Name":"mysql","Version":"e5309aed88f5cc662bc77988a31874461f7c4fb8","Fingerprint":"e5309aed88f5cc662bc77988a31874461f7c4fb8","SHA1":"cc2cb926b6a644d2692700c5ae3bd4f5b84dcb82","TarPath":"/mnt/tmp/tar-CmdExtractor550036964/packages/mysql.tgz","Dependencies":null},{"Name":"ruby","Version":"8c1c0bba2f15f89e3129213e3877dd40e339592f","Fingerprint":"8c1c0bba2f15f89e3129213e3877dd40e339592f","SHA1":"2f894c461afb1586dce78818e3c252f8594736cf","TarPath":"/mnt/tmp/tar-CmdExtractor550036964/packages/ruby.tgz","Dependencies":null}]},{"Name":"redis","Version":"ec27a0b7849863bc160ac54ce667ecacd07fc4cb","Fingerprint":"ec27a0b7849863bc160ac54ce667ecacd07fc4cb","SHA1":"39ca0896434267037575c757e01badb50b15fd0c","TarPath":"/mnt/tmp/tar-CmdExtractor550036964/packages/redis.tgz","Dependencies":null},{"Name":"common","Version":"1324d32dbda40da88aade1e07b226a208602baff","Fingerprint":"1324d32dbda40da88aade1e07b226a208602baff","SHA1":"9407d1d3da30ded1e955acfc8899304aa58a0e78","TarPath":"/mnt/tmp/tar-CmdExtractor550036964/packages/common.tgz","Dependencies":null},{"Name":"libpq","Version":"6aa19afb153dc276924693dc724760664ce61593","Fingerprint":"6aa19afb153dc276924693dc724760664ce61593","SHA1":"7f0f4183dea968373f73091d1c911d5edebc3f25","TarPath":"/mnt/tmp/tar-CmdExtractor550036964/packages/libpq.tgz","Dependencies":null},{"Name":"powerdns","Version":"e41baf8e236b5fed52ba3c33cf646e4b2e0d5a4e","Fingerprint":"e41baf8e236b5fed52ba3c33cf646e4b2e0d5a4e","SHA1":"57a76e65db09002b1b9a277c048767f50f488e31","TarPath":"/mnt/tmp/tar-CmdExtractor550036964/packages/powerdns.tgz","Dependencies":null},{"Name":"mysql","Version":"e5309aed88f5cc662bc77988a31874461f7c4fb8","Fingerprint":"e5309aed88f5cc662bc77988a31874461f7c4fb8","SHA1":"cc2cb926b6a644d2692700c5ae3bd4f5b84dcb82","TarPath":"/mnt/tmp/tar-CmdExtractor550036964/packages/mysql.tgz","Dependencies":null},{"Name":"nats","Version":"6a31c7bb0d5ffa2a9f43c7fd7193193438e20e92","Fingerprint":"6a31c7bb0d5ffa2a9f43c7fd7193193438e20e92","SHA1":"dee70fef63aa35a0d8e1d8fe730b5596904c4a59","TarPath":"/mnt/tmp/tar-CmdExtractor550036964/packages/nats.tgz","Dependencies":[{"Name":"ruby","Version":"8c1c0bba2f15f89e3129213e3877dd40e339592f","Fingerprint":"8c1c0bba2f15f89e3129213e3877dd40e339592f","SHA1":"2f894c461afb1586dce78818e3c252f8594736cf","TarPath":"/mnt/tmp/tar-CmdExtractor550036964/packages/ruby.tgz","Dependencies":null}]}]}
func (r ReleaseVersions) Import(data io.Reader) error {
	rd := bufio.NewReader(data)
	for {
		line, err := rd.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("Reading data: %s", err)
		}

		if len(strings.TrimSpace(line)) == 0 {
			continue
		}

		pieces := strings.SplitN(line, "|", 2)
		if len(pieces) != 2 {
			return fmt.Errorf("Parsing line: '%s'", line)
		}

		pieces[0] = strings.TrimSpace(pieces[0])
		pieces[1] = strings.TrimSpace(pieces[1])

		var k key

		err = json.Unmarshal([]byte(pieces[0]), &k)
		if err != nil {
			return fmt.Errorf("Unmarshaling key: %s", pieces[0])
		}

		if len(k.Source) == 0 || len(k.VersionRaw) == 0 {
			return fmt.Errorf("Invalid key from '%s'", pieces[0])
		}

		fmt.Printf("[%#v] processing\n", k)

		var val value

		err = json.Unmarshal([]byte(pieces[1]), &val)
		if err != nil {
			return fmt.Errorf("Unmarshaling val: %s: %s", pieces[1], err)
		}

		bytes, err := json.MarshalIndent(val, "", "  ")
		if err != nil {
			return fmt.Errorf("Marshaling val: %s", err)
		}

		name := val["Name"].(string)

		releaseDir := filepath.Join(r.ReleasesIndexDir, k.Source, name+"-"+k.VersionRaw)

		err = os.MkdirAll(releaseDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("Making dir: %s", err)
		}

		releaseV1Path := filepath.Join(releaseDir, "release.v1.yml")

		err = ioutil.WriteFile(releaseV1Path, bytes, 0644)
		if err != nil {
			return fmt.Errorf("Writing file: path=%s %s", releaseV1Path, err)
		}
	}

	return nil
}
