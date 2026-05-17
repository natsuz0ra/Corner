import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import test from 'node:test'
import { renderMarkdown } from '../src/utils/markdown'

const projectRoot = resolve(import.meta.dirname, '..')

test('renderMarkdown wraps fenced code blocks with a copy button', () => {
  const html = renderMarkdown('```ts\nconst value = 1\n```', { codeCopyButton: true })

  assert.match(html, /class="markdown-code-block markdown-code-block--single-line"/)
  assert.match(html, /class="markdown-code-copy-btn"/)
  assert.match(html, /aria-label="Copy code"/)
  assert.match(html, /<pre><code class="language-ts">/)
  assert.match(html, /hljs-keyword">const<\/span> value = <span class="hljs-number">1<\/span>/)
})

test('renderMarkdown keeps multiline copy buttons anchored as multiline blocks', () => {
  const html = renderMarkdown('```ts\nconst value = 1\nconsole.log(value)\n```', { codeCopyButton: true })

  assert.match(html, /class="markdown-code-block"/)
  assert.doesNotMatch(html, /markdown-code-block--single-line/)
})

test('renderMarkdown keeps fenced code blocks plain by default', () => {
  const html = renderMarkdown('```ts\nconst value = 1\n```')

  assert.doesNotMatch(html, /markdown-code-block/)
  assert.doesNotMatch(html, /markdown-code-copy-btn/)
  assert.match(html, /<pre><code class="language-ts">/)
})

test('renderMarkdown does not add copy controls for inline code', () => {
  const html = renderMarkdown('Use `const value = 1` inline.', { codeCopyButton: true })

  assert.doesNotMatch(html, /markdown-code-block/)
  assert.doesNotMatch(html, /markdown-code-copy-btn/)
})

test('PlanBlock enables copy controls for fenced code blocks', () => {
  const source = readFileSync(resolve(projectRoot, 'src/components/chat/PlanBlock.vue'), 'utf8')

  assert.match(source, /renderMarkdown\(content,\s*\{\s*codeCopyButton:\s*true\s*\}\)/)
  assert.match(source, /@click="handleMarkdownClick"/)
})

test('markdown code copy trims trailing newlines and uses a minimal icon button', () => {
  const copySource = readFileSync(resolve(projectRoot, 'src/composables/chat/useMarkdownCodeCopy.ts'), 'utf8')
  const markdownStyles = readFileSync(resolve(projectRoot, 'src/styles/markdown.css'), 'utf8')

  assert.match(copySource, /replace\(\s*\/\\n\+\$\/,\s*''\s*\)/)
  assert.match(markdownStyles, /\.markdown-code-copy-btn\s*\{[\s\S]*?width:\s*22px;[\s\S]*?height:\s*22px;[\s\S]*?border:\s*0;[\s\S]*?background:\s*transparent;[\s\S]*?opacity:\s*0\.9;/)
  assert.match(markdownStyles, /\.markdown-code-copy-btn::before\s*\{[\s\S]*?width:\s*14px;[\s\S]*?height:\s*14px;/)
  assert.match(markdownStyles, /\.markdown-code-block--single-line \.markdown-code-copy-btn\s*\{[\s\S]*?top:\s*50%;[\s\S]*?transform:\s*translateY\(-50%\);/)
})
