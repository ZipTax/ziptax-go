package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAddress(t *testing.T) {
	t.Run("valid address with 5-digit zip", func(t *testing.T) {
		result, err := ParseAddress("200 Spectrum Center Dr, Irvine, CA 92618")
		require.NoError(t, err)
		assert.Equal(t, "200 Spectrum Center Dr", result.Line1)
		assert.Equal(t, "", result.Line2)
		assert.Equal(t, "Irvine", result.City)
		assert.Equal(t, "CA", result.State)
		assert.Equal(t, "92618", result.Zip)
		assert.Equal(t, "US", result.CountryCode)
	})

	t.Run("valid address with 9-digit zip", func(t *testing.T) {
		result, err := ParseAddress("200 Spectrum Center Dr, Irvine, CA 92618-1905")
		require.NoError(t, err)
		assert.Equal(t, "200 Spectrum Center Dr", result.Line1)
		assert.Equal(t, "Irvine", result.City)
		assert.Equal(t, "CA", result.State)
		assert.Equal(t, "92618-1905", result.Zip)
		assert.Equal(t, "US", result.CountryCode)
	})

	t.Run("valid address with Minneapolis", func(t *testing.T) {
		result, err := ParseAddress("323 Washington Ave N, Minneapolis, MN 55401-2427")
		require.NoError(t, err)
		assert.Equal(t, "323 Washington Ave N", result.Line1)
		assert.Equal(t, "Minneapolis", result.City)
		assert.Equal(t, "MN", result.State)
		assert.Equal(t, "55401-2427", result.Zip)
		assert.Equal(t, "US", result.CountryCode)
	})

	t.Run("address with extra whitespace", func(t *testing.T) {
		result, err := ParseAddress("  200 Spectrum Center Dr , Irvine ,  CA 92618  ")
		require.NoError(t, err)
		assert.Equal(t, "200 Spectrum Center Dr", result.Line1)
		assert.Equal(t, "Irvine", result.City)
		assert.Equal(t, "CA", result.State)
		assert.Equal(t, "92618", result.Zip)
	})

	t.Run("address with lowercase state", func(t *testing.T) {
		result, err := ParseAddress("200 Spectrum Center Dr, Irvine, ca 92618")
		require.NoError(t, err)
		assert.Equal(t, "CA", result.State)
	})

	t.Run("address with suite maps to Line2", func(t *testing.T) {
		result, err := ParseAddress("200 Spectrum Center Dr, Suite 100, Irvine, CA 92618")
		require.NoError(t, err)
		assert.Equal(t, "200 Spectrum Center Dr", result.Line1)
		assert.Equal(t, "Suite 100", result.Line2)
		assert.Equal(t, "Irvine", result.City)
		assert.Equal(t, "CA", result.State)
		assert.Equal(t, "92618", result.Zip)
	})

	t.Run("address with multiple middle segments joined as Line2", func(t *testing.T) {
		result, err := ParseAddress("200 Spectrum Center Dr, Bldg A, Suite 100, Irvine, CA 92618")
		require.NoError(t, err)
		assert.Equal(t, "200 Spectrum Center Dr", result.Line1)
		assert.Equal(t, "Bldg A, Suite 100", result.Line2)
		assert.Equal(t, "Irvine", result.City)
		assert.Equal(t, "CA", result.State)
		assert.Equal(t, "92618", result.Zip)
	})

	t.Run("address with apartment number", func(t *testing.T) {
		result, err := ParseAddress("123 Main St, Apt 4B, New York, NY 10001")
		require.NoError(t, err)
		assert.Equal(t, "123 Main St", result.Line1)
		assert.Equal(t, "Apt 4B", result.Line2)
		assert.Equal(t, "New York", result.City)
		assert.Equal(t, "NY", result.State)
		assert.Equal(t, "10001", result.Zip)
	})

	t.Run("three segments has no Line2", func(t *testing.T) {
		result, err := ParseAddress("200 Spectrum Center Dr, Irvine, CA 92618")
		require.NoError(t, err)
		assert.Equal(t, "200 Spectrum Center Dr", result.Line1)
		assert.Equal(t, "", result.Line2)
		assert.Equal(t, "Irvine", result.City)
	})

	t.Run("empty address", func(t *testing.T) {
		_, err := ParseAddress("")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})

	t.Run("whitespace only address", func(t *testing.T) {
		_, err := ParseAddress("   ")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})

	t.Run("address with no commas", func(t *testing.T) {
		_, err := ParseAddress("200 Spectrum Center Dr Irvine CA 92618")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "at least 3 comma-separated segments")
	})

	t.Run("address with only one comma", func(t *testing.T) {
		_, err := ParseAddress("200 Spectrum Center Dr, Irvine CA 92618")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "at least 3 comma-separated segments")
	})

	t.Run("empty street segment", func(t *testing.T) {
		_, err := ParseAddress(", Irvine, CA 92618")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "street address")
		assert.Contains(t, err.Error(), "cannot be empty")
	})

	t.Run("empty city segment", func(t *testing.T) {
		_, err := ParseAddress("200 Spectrum Center Dr, , CA 92618")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "city")
		assert.Contains(t, err.Error(), "cannot be empty")
	})

	t.Run("empty middle segment skipped in Line2", func(t *testing.T) {
		// Empty middle segments between line1 and city are skipped
		result, err := ParseAddress("200 Spectrum Center Dr, , Suite 100, Irvine, CA 92618")
		require.NoError(t, err)
		assert.Equal(t, "200 Spectrum Center Dr", result.Line1)
		assert.Equal(t, "Suite 100", result.Line2)
		assert.Equal(t, "Irvine", result.City)
	})

	t.Run("empty state zip segment", func(t *testing.T) {
		_, err := ParseAddress("200 Spectrum Center Dr, Irvine, ")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "state and zip")
		assert.Contains(t, err.Error(), "cannot be empty")
	})

	t.Run("state zip without zip code", func(t *testing.T) {
		_, err := ParseAddress("200 Spectrum Center Dr, Irvine, CA")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "last segment must contain exactly state and zip code")
	})

	t.Run("invalid state length", func(t *testing.T) {
		_, err := ParseAddress("200 Spectrum Center Dr, Irvine, CAL 92618")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "2-letter abbreviation")
	})

	t.Run("invalid zip code format", func(t *testing.T) {
		_, err := ParseAddress("200 Spectrum Center Dr, Irvine, CA 1234")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid zip code")
	})

	t.Run("zip code with letters", func(t *testing.T) {
		_, err := ParseAddress("200 Spectrum Center Dr, Irvine, CA 9261A")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid zip code")
	})

	t.Run("extra fields in state zip segment", func(t *testing.T) {
		_, err := ParseAddress("200 Spectrum Center Dr, Irvine, CA 92618 USA")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "last segment must contain exactly state and zip code")
	})

	t.Run("extra fields in state zip segment with split zip", func(t *testing.T) {
		_, err := ParseAddress("323 Washington Ave N, Minneapolis, MN 55401 2427")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "last segment must contain exactly state and zip code")
	})
}
