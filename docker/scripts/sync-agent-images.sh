#!/usr/bin/env bash
# Promote: coin-jenkins-agents/catalog.yaml → coin-lib/resources/images.yaml (+ images-local.yaml).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
CATALOG="${ROOT}/../coin-jenkins-agents/catalog.yaml"
IMAGES="${ROOT}/../coin-lib/resources/images.yaml"
LOCAL="${ROOT}/images-local.yaml"

ruby - "$CATALOG" "$IMAGES" "$LOCAL" <<'RUBY'
require 'yaml'

catalog_path, images_path, local_path = ARGV
catalog = YAML.load_file(catalog_path)

def load_yaml(path)
  File.exist?(path) ? YAML.load_file(path) : {}
end

def save_yaml(path, data)
  File.write(path, data.to_yaml)
end

def image_name(raw)
  raw.start_with?('coin/') ? raw : "coin/#{raw}"
end

def promote(target, catalog, registry_prefix: nil)
  (catalog['stacks'] || {}).each do |stack, runtimes|
    target['stacks'] ||= {}
    target['stacks'][stack] ||= {}
    runtimes.each do |runtime, entry|
      tag = entry['tag'].to_s.empty? ? "#{runtime}-r#{entry['rev'] || 0}" : entry['tag']
      name = image_name(entry['image'] || stack)
      ref = if registry_prefix
              short = name.split('/', 2)[1]
              "#{registry_prefix}/#{short}:#{tag}"
            else
              "#{name}:#{tag}"
            end
      target['stacks'][stack][runtime] = {
        'image' => ref,
        'digest' => entry['digest'].to_s,
        'rev' => entry['rev'] || 0,
      }
    end
  end
end

images = load_yaml(images_path)
promote(images, catalog)
save_yaml(images_path, images)

if File.exist?(local_path)
  local = load_yaml(local_path)
  registry = catalog.dig('registry', 'default') || 'registry:5000/coin'
  host_reg = registry.sub('registry:5000', 'localhost:5050')
  promote(local, catalog, registry_prefix: host_reg)
  save_yaml(local_path, local)
  puts 'synced catalog -> images.yaml, images-local.yaml'
else
  puts 'synced catalog -> images.yaml'
end
RUBY
