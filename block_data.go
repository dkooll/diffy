package diffy

import (
	"fmt"
	"strings"

	"golang.org/x/exp/slices"
)

// NewBlockData creates a new empty BlockData
func NewBlockData() BlockData {
	return BlockData{
		Properties:    make(map[string]bool),
		StaticBlocks:  make(map[string]*ParsedBlock),
		DynamicBlocks: make(map[string]*ParsedBlock),
		IgnoreChanges: []string{},
	}
}

// ParseAttributes extracts attributes from a body into properties
func (bd *BlockData) ParseAttributes(body *Body) {
	for name := range body.Attributes {
		bd.Properties[name] = true
	}
}

// ParseBlocks processes all blocks in a body
func (bd *BlockData) ParseBlocks(body *Body) {
	var directIgnoreChanges []string

	// In a real implementation, this would extract ignore_changes from lifecycle blocks
	// directIgnoreChanges = extractLifecycleIgnoreChangesFromAST(body)

	if len(directIgnoreChanges) > 0 {
		bd.IgnoreChanges = append(bd.IgnoreChanges, directIgnoreChanges...)
	}

	for _, block := range body.Blocks {
		switch block.Type {
		case "lifecycle":
			bd.parseLifecycle(block.Body)
		case "dynamic":
			if len(block.Labels) == 1 {
				bd.parseDynamicBlock(block.Body, block.Labels[0])
			}
		default:
			parsed := ParseBody(block.Body)
			bd.StaticBlocks[block.Type] = parsed
		}
	}
}

// parseLifecycle extracts ignore_changes from a lifecycle block
func (bd *BlockData) parseLifecycle(body *Body) {
	// In a real implementation, this would extract the ignore_changes attribute
	// For the sample, we just demonstrate the concept
	if ignoreChanges, ok := body.Attributes["ignore_changes"]; ok {
		if list, ok := ignoreChanges.([]string); ok {
			bd.IgnoreChanges = append(bd.IgnoreChanges, list...)
		} else if str, ok := ignoreChanges.(string); ok && str == "all" {
			bd.IgnoreChanges = append(bd.IgnoreChanges, "*all*")
		}
	}
}

// parseDynamicBlock processes a dynamic block
func (bd *BlockData) parseDynamicBlock(body *Body, name string) {
	contentBlock := findContentBlock(body)
	parsed := ParseBody(contentBlock)
	if existing := bd.DynamicBlocks[name]; existing != nil {
		mergeBlocks(existing, parsed)
	} else {
		bd.DynamicBlocks[name] = parsed
	}
}

// Validate recursively validates a block against its schema
func (bd *BlockData) Validate(
	resourceType, path string,
	schema *SchemaBlock,
	parentIgnore []string,
	findings *[]ValidationFinding,
) {
	if schema == nil {
		return
	}

	ignore := slices.Clone(parentIgnore)
	ignore = append(ignore, bd.IgnoreChanges...)

	bd.validateAttributes(resourceType, path, schema, ignore, findings)
	bd.validateBlocks(resourceType, path, schema, ignore, findings)
}

// validateAttributes checks for missing required attributes
func (bd *BlockData) validateAttributes(
	resType, path string,
	schema *SchemaBlock,
	ignore []string,
	findings *[]ValidationFinding,
) {
	for name, attr := range schema.Attributes {
		if name == "id" {
			continue
		}

		if attr.Computed && !attr.Optional && !attr.Required {
			continue
		}

		if isIgnored(ignore, name) {
			continue
		}

		if !bd.Properties[name] {
			*findings = append(*findings, ValidationFinding{
				ResourceType: resType,
				Path:         path,
				Name:         name,
				Required:     attr.Required,
				IsBlock:      false,
			})
		}
	}
}

// validateBlocks checks for missing required blocks
func (bd *BlockData) validateBlocks(
	resType, path string,
	schema *SchemaBlock,
	ignore []string,
	findings *[]ValidationFinding,
) {
	for name, blockType := range schema.BlockTypes {
		if name == "timeouts" || isIgnored(ignore, name) {
			continue
		}

		static := bd.StaticBlocks[name]
		dynamic := bd.DynamicBlocks[name]

		if static == nil && dynamic == nil {
			*findings = append(*findings, ValidationFinding{
				ResourceType: resType,
				Path:         path,
				Name:         name,
				Required:     blockType.MinItems > 0,
				IsBlock:      true,
			})
			continue
		}

		var target *ParsedBlock
		if static != nil {
			target = static
		} else {
			target = dynamic
		}

		newPath := fmt.Sprintf("%s.%s", path, name)
		target.Data.Validate(resType, newPath, blockType.Block, ignore, findings)
	}
}

// ParseBody parses a Body into a ParsedBlock
func ParseBody(body *Body) *ParsedBlock {
	bd := NewBlockData()
	blk := &ParsedBlock{Data: bd}
	bd.ParseAttributes(body)
	bd.ParseBlocks(body)
	return blk
}

// Helper functions

// findContentBlock finds the content block within a dynamic block
func findContentBlock(body *Body) *Body {
	for _, b := range body.Blocks {
		if b.Type == "content" {
			return b.Body
		}
	}
	return body
}

// isIgnored checks if a property should be ignored
func isIgnored(ignore []string, name string) bool {
	if slices.Contains(ignore, "*all*") {
		return true
	}

	for _, item := range ignore {
		if strings.EqualFold(item, name) {
			return true
		}
	}
	return false
}

// mergeBlocks combines two parsed blocks
func mergeBlocks(dest, src *ParsedBlock) {
	for k := range src.Data.Properties {
		dest.Data.Properties[k] = true
	}

	for k, v := range src.Data.StaticBlocks {
		if existing, ok := dest.Data.StaticBlocks[k]; ok {
			mergeBlocks(existing, v)
		} else {
			dest.Data.StaticBlocks[k] = v
		}
	}

	for k, v := range src.Data.DynamicBlocks {
		if existing, ok := dest.Data.DynamicBlocks[k]; ok {
			mergeBlocks(existing, v)
		} else {
			dest.Data.DynamicBlocks[k] = v
		}
	}

	dest.Data.IgnoreChanges = append(dest.Data.IgnoreChanges, src.Data.IgnoreChanges...)
}
