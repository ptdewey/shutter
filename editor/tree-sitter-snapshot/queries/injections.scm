; Language injections for snapshot body content
;
; Explicit content_type field in header takes precedence if present.
; Otherwise, JSON objects and arrays are detected by heuristic.

; Explicit content_type field takes precedence
((front_matter
   (field
     (field_name) @_name
     (field_value) @injection.language))
  (body) @injection.content
  (#eq? @_name "content_type"))

; JSON object: starts with {
((body) @injection.content
  (#match? @injection.content "^\\s*\\{")
  (#set! injection.language "json"))

; JSON array: starts with [
((body) @injection.content
  (#match? @injection.content "^\\s*\\[")
  (#set! injection.language "json"))

