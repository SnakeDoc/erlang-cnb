package fakes

import "sync"

type Installer struct {
	BuildDownloadURLCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Arch          string
			UbuntuVersion string
			Version       string
		}
		Returns struct {
			String string
		}
		Stub func(string, string, string) string
	}
	InstallCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Url       string
			LayerPath string
		}
		Returns struct {
			Error error
		}
		Stub func(string, string) error
	}
}

func (f *Installer) BuildDownloadURL(param1 string, param2 string, param3 string) string {
	f.BuildDownloadURLCall.mutex.Lock()
	defer f.BuildDownloadURLCall.mutex.Unlock()
	f.BuildDownloadURLCall.CallCount++
	f.BuildDownloadURLCall.Receives.Arch = param1
	f.BuildDownloadURLCall.Receives.UbuntuVersion = param2
	f.BuildDownloadURLCall.Receives.Version = param3
	if f.BuildDownloadURLCall.Stub != nil {
		return f.BuildDownloadURLCall.Stub(param1, param2, param3)
	}
	return f.BuildDownloadURLCall.Returns.String
}
func (f *Installer) Install(param1 string, param2 string) error {
	f.InstallCall.mutex.Lock()
	defer f.InstallCall.mutex.Unlock()
	f.InstallCall.CallCount++
	f.InstallCall.Receives.Url = param1
	f.InstallCall.Receives.LayerPath = param2
	if f.InstallCall.Stub != nil {
		return f.InstallCall.Stub(param1, param2)
	}
	return f.InstallCall.Returns.Error
}
