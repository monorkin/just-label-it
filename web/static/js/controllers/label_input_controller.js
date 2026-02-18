(() => {
  const { Controller } = Stimulus

  class LabelInputController extends Controller {
    static targets = ["input", "suggestions", "tags"]
    static values = { url: String }

    connect() {
      this._activeIndex = -1
      this._suggestions = []
      this._searchTimeout = null
    }

    disconnect() {
      clearTimeout(this._searchTimeout)
    }

    search() {
      clearTimeout(this._searchTimeout)
      const query = this.inputTarget.value.trim()
      if (!query) {
        this._hideSuggestions()
        return
      }

      this._searchTimeout = setTimeout(() => {
        this._fetchSuggestions(query)
      }, 200)
    }

    keydown(event) {
      if (event.key === "Enter") {
        event.preventDefault()
        if (this._activeIndex >= 0 && this._activeIndex < this._suggestions.length) {
          this._addLabel(this._suggestions[this._activeIndex].name)
        } else if (this.inputTarget.value.trim()) {
          this._addLabel(this.inputTarget.value.trim())
        }
      } else if (event.key === "ArrowDown") {
        event.preventDefault()
        this._moveSelection(1)
      } else if (event.key === "ArrowUp") {
        event.preventDefault()
        this._moveSelection(-1)
      } else if (event.key === "Escape") {
        this._hideSuggestions()
      }
    }

    removeLabel(event) {
      const labelId = event.currentTarget.dataset.labelId
      const tag = event.currentTarget.closest(".label-tag")
      const url = this.urlValue
      if (!url) return

      fetch(`${url}/${labelId}`, { method: "DELETE" }).then(response => {
        if (response.ok && tag) {
          tag.remove()
        }
      })
    }

    _addLabel(name) {
      const url = this.urlValue
      if (!url) return

      fetch(url, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name })
      })
        .then(r => r.json())
        .then(label => {
          this._appendTag(label)
          this.inputTarget.value = ""
          this._hideSuggestions()
        })
    }

    _appendTag(label) {
      // Don't add duplicate tags.
      if (this.tagsTarget.querySelector(`[data-label-id="${label.ID}"]`)) return

      const span = document.createElement("span")
      span.className = "label-tag"
      span.dataset.labelId = label.ID
      span.innerHTML = `${this._escapeHtml(label.Name)} <button class="label-remove" data-action="click->label-input#removeLabel" data-label-id="${label.ID}">&times;</button>`
      this.tagsTarget.appendChild(span)
    }

    _fetchSuggestions(query) {
      fetch(`/api/labels?q=${encodeURIComponent(query)}`)
        .then(r => r.json())
        .then(labels => {
          this._suggestions = labels
          this._activeIndex = -1
          this._renderSuggestions(query)
        })
    }

    _renderSuggestions(query) {
      if (this._suggestions.length === 0) {
        this._hideSuggestions()
        return
      }

      const html = this._suggestions.map((label, i) => {
        const highlighted = this._highlight(label.Name, query)
        const cls = i === this._activeIndex ? "label-suggestion active" : "label-suggestion"
        return `<div class="${cls}" data-index="${i}" onmousedown="event.preventDefault()" onclick="this.closest('[data-controller~=label-input]').__stimulusLabelClick(${i})">${highlighted}</div>`
      }).join("")

      this.suggestionsTarget.innerHTML = html
      this.suggestionsTarget.style.display = "block"

      // Store click handler on element for the inline onclick.
      this.element.__stimulusLabelClick = (index) => {
        this._addLabel(this._suggestions[index].name)
      }
    }

    _hideSuggestions() {
      this.suggestionsTarget.style.display = "none"
      this._suggestions = []
      this._activeIndex = -1
    }

    _moveSelection(direction) {
      if (this._suggestions.length === 0) return
      this._activeIndex = Math.max(-1, Math.min(this._suggestions.length - 1, this._activeIndex + direction))

      const items = this.suggestionsTarget.querySelectorAll(".label-suggestion")
      items.forEach((el, i) => {
        el.classList.toggle("active", i === this._activeIndex)
      })
    }

    _highlight(text, query) {
      const idx = text.toLowerCase().indexOf(query.toLowerCase())
      if (idx === -1) return this._escapeHtml(text)
      const before = text.slice(0, idx)
      const match = text.slice(idx, idx + query.length)
      const after = text.slice(idx + query.length)
      return `${this._escapeHtml(before)}<mark>${this._escapeHtml(match)}</mark>${this._escapeHtml(after)}`
    }

    _escapeHtml(str) {
      const div = document.createElement("div")
      div.textContent = str
      return div.innerHTML
    }
  }

  window.StimulusApp.register("label-input", LabelInputController)
})()
