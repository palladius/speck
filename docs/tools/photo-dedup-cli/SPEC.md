---
speck_version: "0.1"
mode: oneshot
idea_file: input_prompt.md
inspiration_dir: discussion.md
inspiration_files:
    - .
created_at: "2026-07-08T09:48:45Z"
model: gemini-flash-latest
tokens:
    prompt: 1559
    output: 1534
    total: 4277
---

# Deterministic & LLM-Powered Photo Deduplication CLI

## Problem Statement

Photo collections grow rapidly across multiple locations, including local drives, Network Attached Storage (NAS), and cloud platforms like Google Cloud Storage (GCS). Over time, these collections accumulate redundant files, split across two distinct categories:
1. **Exact duplicates**: Identical byte-for-byte copies that reside under different file paths or names.
2. **Near-duplicates**: Photos taken in a burst, the same image saved in multiple resolutions, or the same photo with slightly different descriptive filenames vs. generic IDs.

Existing deduplication utilities are either restricted to local filesystems, lacks native integration with cloud storage providers like GCS, or do not offer intelligent semantic or visual evaluation for near-duplicate cases. Users need a reliable tool that is blazingly fast and deterministic for exact duplicates, yet smart enough to run semantic visual inspections when prompted.

## Goals

- **Unified Backend Interface**: Support both local disks/NAS and Google Cloud Storage (GCS) through a common storage abstraction layer.
- **Fast, Deterministic Exact Matching**: Find byte-for-byte duplicates quickly using size-matching followed by SHA-256 hashing.
- **Boundary Protection ("World" Isolation)**: Ensure deduplication tasks run strictly within a single storage target (a local path or a GCS bucket path) to prevent accidental loss across different systems.
- **Smart programmatic selection**: Determine the "winner" of an exact duplicate group using a strict rule hierarchy (oldest creation date, then shortest filename path).
- **Non-destructive quarantine**: Safely isolate duplicate photos in a `.trash/` directory at the root of the target "world" instead of executing immediate deletions.
- **Near-duplicate inspection (`--llm-inspect`)**: Leverage lightweight perceptual hashing (pHash) to flag visually similar candidates, and query a multimodal LLM to evaluate complex metadata (such as prioritizing descriptive filenames over generic IDs like `IMG_0412.jpg`) and visually verify true near-duplicates.

## Non-Goals

- Cross-system synchronization or cross-world deduplication (e.g., treating GCS as a master mirror to delete local files, or vice-versa).
- Active background system daemons for real-time filesystem file-watcher monitoring.
- Dedicated graphical user interface (GUI); this is strictly a CLI utility.
- Deduplication of non-media files (e.g., documents, raw database files).

## Technical Plan / Approach

### 1. Storage Abstraction Layer
We will define a common interface `StorageProvider` to handle different targets:

```go
type FileMetadata struct {
    Path        string
    Size        int64
    CreatedAt   time.Time
    Hash        string
}

type StorageProvider interface {
    ListFiles(prefix string) ([]FileMetadata, error)
    ReadBlock(path string, offset int64, size int64) ([]byte, error)
    Move(src string, dest string) error
}
```
We will implement `LocalDiskProvider` and `GCSProvider` matching this interface.

### 2. Exact Duplication Pipeline
- **Group by Size**: List all files and map size to list of files. Paths with unique sizes are skipped instantly.
- **Block-based Hashing**: For files sharing the exact same size, read and compare the first and last 4KB. If those match, compute the full SHA-256 hash to confirm identity.
- **Resolution Rules**: For exact matches, resolve the "winner" systematically:
  1. Oldest creation/modification time wins.
  2. If times are identical, the file with the shortest absolute path length wins.
- **Quarantine Execution**: Move "loser" files into a `.trash/` directory structure under the scanned root.

### 3. Smart Near-Duplication (`--llm-inspect` Mode)
- Run a preliminary fast perceptual hash (pHash) scan on downscaled local proxies of images to gather potential near-duplicate clusters.
- For flagged near-duplicate clusters, extract metadata (filename, date, resolution) and run visual comparison prompts via a cost-effective multimodal LLM (e.g., Gemini Flash or GPT-4o-mini).
- The LLM assesses which file is the superior original based on image resolution, visual quality, and descriptive filename strength (e.g., preferring `yellowstone_trip_cabin.jpg` over `IMG_90812_copy.jpg`).
- Quarantine near-duplicates only after presenting a clean confirmation step or moving them straight to `.trash/` if run in a high-confidence automated configuration.

## Alternatives Considered

- **Using system tools (`fdupes`, `czkawka`)**: Dismissed because they don't support GCS natively and cannot handle remote network latencies gracefully without downloading entire datasets. They also lack integration with LLM inspection models.
- **Heavyweight Local Vector Databases**: Generating high-dimensional visual embeddings for every image and storing them in a local vector database. This is too heavy for a quick CLI tool that handles a few thousand local photos. Perceptual hashing (pHash) combined with downstream LLM evaluation provides a much lighter footprint.

## Implementation Plan

- **Phase 1: Foundation (CLI & Local Provider)**: Build CLI skeleton, design `StorageProvider` interface, and build the `LocalDiskProvider` along with deterministic exact-duplicate matching.
- **Phase 2: Cloud Support (GCS Provider)**: Implement `GCSProvider` and add test configurations using a GCS emulator or live staging bucket.
- **Phase 3: Quarantine Logic & Resolution Engine**: Implement path resolution engines (oldest-wins, shortest-path-wins) and robust quarantine operations to safely move assets into `.trash/` folders without data loss.
- **Phase 4: Perceptual Hashing & LLM-Inspect**: Integrate the pHash computation library, establish the API adapter for multimodal LLM parsing, and write prompt templates for near-duplicate metadata/visual analysis.
- **Phase 5: Output, Reporting, and Interactive Mode**: Implement human-readable console outputs detailing what was quarantined and allow interactive user confirmation in dry-run scenarios.

## Open Questions

- **GCS Read Costs**: To calculate SHA-256 hashes of matching file sizes, we must read the entire file stream. For GCS, this incurs egress/API access charges. Should we implement an optional persistent local database to cache GCS object metadata and previously calculated hashes?
- **Multimodal LLM Payload size**: What is the best strategy for feeding near-duplicate images to the LLM when operating on GCS? Passing signed URLs directly to the API instead of downloading and sending base64 payloads would significantly reduce runtime latency and local network overhead.
