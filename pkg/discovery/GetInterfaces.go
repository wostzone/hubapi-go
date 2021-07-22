package discovery

import (
	"net"

	"github.com/sirupsen/logrus"
)

// Get a list of active network interfaces excluding the loopback interface
//  address to only return the interface that serves the given IP address
func GetInterfaces(address string) ([]net.Interface, error) {
	result := make([]net.Interface, 0)
	ip := net.ParseIP(address)

	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		// ignore interfaces without address
		if err == nil {
			for _, a := range addrs {
				switch v := a.(type) {
				case *net.IPAddr:
					result = append(result, iface)
					logrus.Infof("GetInterfaces: Found: Interface%s", v.String())

				case *net.IPNet:
					ifNet := a.(*net.IPNet)
					hasIP := ifNet.Contains(ip)

					// ignore loopback interface
					if hasIP && !a.(*net.IPNet).IP.IsLoopback() {
						result = append(result, iface)
						logrus.Infof("GetInterfaces: Found network %v : %s [%v/%v]\n", iface.Name, v, v.IP, v.Mask)
					}
				}
			}
		}
	}
	return result, nil
}
