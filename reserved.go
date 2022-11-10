package mmdbwriter

// These were taken from the Perl writer.
//
// https://www.iana.org/assignments/iana-ipv4-special-registry/iana-ipv4-special-registry.xhtml
var reservedNetworksIPv4 = []string{
	"0.0.0.0/8",
}

// https://www.iana.org/assignments/iana-ipv6-special-registry/iana-ipv6-special-registry.xhtml
var reservedNetworksIPv6 = []string{
	// ::/128 and ::1/128 are reserved under IPv6 but these are already
	// covered under 0.0.0.0/8.
	//
	// ::ffff:0:0/96 - IPv4 mapped addresses. We treat it specially with the
	// `alias_ipv6_to_ipv4' option.
	//
	// 64:ff9b::/96 - well known prefix mapping, covered by alias_ipv6_to_ipv4
	//
	// TODO(wstorey@maxmind.com): 64:ff9b:1::/48 should be in
	// alias_ipv6_to_ipv4?

	"100::/64",

	// 2001::/23 is reserved. We include all of it here other than 2001::/32
	// as it is Teredo which is globally routable.
	"2001:1::/32",
	"2001:2::/31",
	"2001:4::/30",
	"2001:8::/29",
	"2001:10::/28",
	"2001:20::/27",
	"2001:40::/26",
	"2001:80::/25",
	"2001:100::/24",

	"2001:db8::/32",
	// 2002::/16 - 6to4, part of alias_ipv6_to_ipv4
	// 2620:4f:8000::/48 is routable I believe
	"fc00::/7",
	"fe80::/10",
	// Multicast
	"ff00::/8",
}
