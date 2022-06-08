package limits

import "github.com/gophercloud/gophercloud"

const rootPath = "limits"

func rootURL(client *gophercloud.ServiceClient) string {
	return client.ServiceURL(rootPath)
}

func resourceURL(client *gophercloud.ServiceClient, id string) string {
	return client.ServiceURL(rootPath, id)
}
