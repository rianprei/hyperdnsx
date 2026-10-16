package dnsx

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolverMode_Default(t *testing.T) {
	require.Equal(t, ResolverModeFirstSuccess, ResolverMode("first-success"))
	require.Equal(t, ResolverModeAll, ResolverMode("all"))
	require.Equal(t, ResolverModeQuorum, ResolverMode("quorum"))
	require.Equal(t, ResolverModeCompare, ResolverMode("compare"))
}

func TestResolverMode_IsValid(t *testing.T) {
	require.True(t, ResolverModeFirstSuccess.IsValid())
	require.True(t, ResolverModeAll.IsValid())
	require.True(t, ResolverModeQuorum.IsValid())
	require.True(t, ResolverModeCompare.IsValid())
	require.False(t, ResolverMode("bogus").IsValid())
	require.False(t, ResolverMode("").IsValid())
}

func TestResolverMode_String(t *testing.T) {
	require.Equal(t, "first-success", ResolverModeFirstSuccess.String())
	require.Equal(t, "all", ResolverModeAll.String())
	require.Equal(t, "quorum", ResolverModeQuorum.String())
	require.Equal(t, "compare", ResolverModeCompare.String())
}

func TestProtocolFromResolver(t *testing.T) {
	require.Equal(t, "udp", protocolFromResolver("udp:1.1.1.1:53"))
	require.Equal(t, "tcp", protocolFromResolver("tcp:8.8.8.8:53"))
	require.Equal(t, "doh", protocolFromResolver("doh:https://dns.google/dns-query"))
	require.Equal(t, "dot", protocolFromResolver("dot:1.1.1.1:853"))
	require.Equal(t, "udp", protocolFromResolver("1.1.1.1"))
}
