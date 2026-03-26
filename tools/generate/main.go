// tools/generate generates Terraform HCL files from the current LuckPerms REST API state.
//
// Usage:
//
//	go run ./tools/generate --url http://localhost:8080 --output ./generated/
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/digitaldrugstech/terraform-provider-luckperms/internal/client"
	"github.com/digitaldrugstech/terraform-provider-luckperms/internal/util"
)

func main() {
	url := flag.String("url", "http://localhost:8080", "LuckPerms REST API base URL")
	output := flag.String("output", "./generated/", "Output directory for generated .tf files")
	apiKey := flag.String("api-key", os.Getenv("LUCKPERMS_API_KEY"), "API key (default: LUCKPERMS_API_KEY env)")
	insecure := flag.Bool("insecure", false, "Skip TLS certificate verification")
	flag.Parse()

	c := client.New(*url, *apiKey, 30*time.Second, *insecure)
	ctx := context.Background()

	if err := os.MkdirAll(*output, 0755); err != nil {
		log.Fatalf("creating output directory: %v", err)
	}

	if err := generateGroups(ctx, c, *output); err != nil {
		log.Fatalf("generating groups: %v", err)
	}

	if err := generateTracks(ctx, c, *output); err != nil {
		log.Fatalf("generating tracks: %v", err)
	}

	log.Printf("done — files written to %s", *output)
}

// generateGroups fetches all groups and writes groups.tf and group_nodes.tf.
func generateGroups(ctx context.Context, c *client.Client, outputDir string) error {
	names, err := c.GetGroups(ctx)
	if err != nil {
		return fmt.Errorf("listing groups: %w", err)
	}
	sort.Strings(names)

	var groupsBuf strings.Builder
	var nodesBuf strings.Builder

	for _, name := range names {
		nodes, err := c.GetGroupNodes(ctx, name)
		if err != nil {
			return fmt.Errorf("fetching nodes for group %q: %w", name, err)
		}

		metaNodes, permNodes := util.SplitNodes(nodes)
		meta := util.ParseMetaNodes(metaNodes)
		permNodes = util.NormalizeNodes(permNodes)

		writeGroupResource(&groupsBuf, name, meta)
		if len(permNodes) > 0 {
			writeGroupNodesResource(&nodesBuf, name, permNodes)
		}
	}

	if err := writeFile(filepath.Join(outputDir, "groups.tf"), groupsBuf.String()); err != nil {
		return err
	}
	if err := writeFile(filepath.Join(outputDir, "group_nodes.tf"), nodesBuf.String()); err != nil {
		return err
	}

	log.Printf("wrote %d groups", len(names))
	return nil
}

// generateTracks fetches all tracks and writes tracks.tf.
func generateTracks(ctx context.Context, c *client.Client, outputDir string) error {
	names, err := c.GetTracks(ctx)
	if err != nil {
		return fmt.Errorf("listing tracks: %w", err)
	}
	sort.Strings(names)

	var buf strings.Builder

	for _, name := range names {
		track, err := c.GetTrack(ctx, name)
		if err != nil {
			return fmt.Errorf("fetching track %q: %w", name, err)
		}
		writeTrackResource(&buf, track)
	}

	if err := writeFile(filepath.Join(outputDir, "tracks.tf"), buf.String()); err != nil {
		return err
	}

	log.Printf("wrote %d tracks", len(names))
	return nil
}

// writeGroupResource appends a luckperms_group resource block to buf.
func writeGroupResource(buf *strings.Builder, name string, meta util.MetaAttrs) {
	id := sanitizeID(name)
	fmt.Fprintf(buf, "resource %q %q {\n", "luckperms_group", id)
	fmt.Fprintf(buf, "  name   = %q\n", name)
	if meta.HasDisplayName {
		fmt.Fprintf(buf, "  display_name = %q\n", meta.DisplayName)
	}
	fmt.Fprintf(buf, "  weight = %d\n", meta.Weight)
	if meta.HasPrefix {
		fmt.Fprintf(buf, "  prefix = %q\n", meta.Prefix)
	}
	if meta.HasSuffix {
		fmt.Fprintf(buf, "  suffix = %q\n", meta.Suffix)
	}
	fmt.Fprintf(buf, "}\n\n")
}

// writeGroupNodesResource appends a luckperms_group_nodes resource block to buf.
func writeGroupNodesResource(buf *strings.Builder, groupName string, nodes []client.Node) {
	id := sanitizeID(groupName)
	fmt.Fprintf(buf, "resource %q %q {\n", "luckperms_group_nodes", id)
	fmt.Fprintf(buf, "  group = luckperms_group.%s.name\n", sanitizeID(groupName))
	fmt.Fprintln(buf)

	for _, n := range nodes {
		writeNodeBlock(buf, n, "  ")
	}

	fmt.Fprintf(buf, "}\n\n")
}

// writeNodeBlock appends a single node {} block to buf.
func writeNodeBlock(buf *strings.Builder, n client.Node, indent string) {
	fmt.Fprintf(buf, "%snode {\n", indent)
	fmt.Fprintf(buf, "%s  key   = %q\n", indent, n.Key)
	if !n.Value {
		fmt.Fprintf(buf, "%s  value = false\n", indent)
	}
	if n.Expiry != nil {
		fmt.Fprintf(buf, "%s  expiry = %d\n", indent, *n.Expiry)
	}
	for _, ctx := range n.Context {
		fmt.Fprintf(buf, "%s  context {\n", indent)
		fmt.Fprintf(buf, "%s    key   = %q\n", indent, ctx.Key)
		fmt.Fprintf(buf, "%s    value = %q\n", indent, ctx.Value)
		fmt.Fprintf(buf, "%s  }\n", indent)
	}
	fmt.Fprintf(buf, "%s}\n", indent)
}

// writeTrackResource appends a luckperms_track resource block to buf.
func writeTrackResource(buf *strings.Builder, track *client.Track) {
	id := sanitizeID(track.Name)
	fmt.Fprintf(buf, "resource %q %q {\n", "luckperms_track", id)
	fmt.Fprintf(buf, "  name = %q\n", track.Name)
	fmt.Fprintf(buf, "  groups = [\n")
	for _, g := range track.Groups {
		fmt.Fprintf(buf, "    luckperms_group.%s.name,\n", sanitizeID(g))
	}
	fmt.Fprintf(buf, "  ]\n")
	fmt.Fprintf(buf, "}\n\n")
}

var nonIdentRe = regexp.MustCompile(`[^a-zA-Z0-9_]`)

// sanitizeID converts a LuckPerms name to a valid Terraform resource identifier.
// Hyphens become underscores; other non-identifier characters are dropped.
func sanitizeID(name string) string {
	s := nonIdentRe.ReplaceAllString(name, "_")
	if len(s) > 0 && s[0] >= '0' && s[0] <= '9' {
		s = "_" + s
	}
	return s
}

// writeFile writes content to path, creating or truncating the file.
func writeFile(path, content string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating %s: %w", path, err)
	}
	defer f.Close()
	if _, err := f.WriteString(content); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return nil
}
