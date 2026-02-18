(() => {
  const { Controller } = Stimulus

  class LabelInputController extends Controller {
    static targets = ["input", "suggestions", "tags"]
    static values = { url: String }

    #activeIndex = -1
    #suggestions = []
    #searchTimeout = null

    disconnect() {
      clearTimeout(this.#searchTimeout)
    }

    search() {
      clearTimeout(this.#searchTimeout)
      const query = this.inputTarget.value.trim()
      if (!query) {
        this.#hideSuggestions()
        return
      }

      this.#searchTimeout = setTimeout(() => {
        this.#fetchSuggestions(query)
      }, 200)
    }

    submitOrNavigate(event) {
      if (event.key === "Enter") {
        event.preventDefault()
        if (this.#activeIndex >= 0 && this.#activeIndex < this.#suggestions.length) {
          this.#addLabel(this.#suggestions[this.#activeIndex].name)
        } else if (this.inputTarget.value.trim()) {
          this.#addLabel(this.inputTarget.value.trim())
        }
      } else if (event.key === "ArrowDown") {
        event.preventDefault()
        this.#moveSelection(1)
      } else if (event.key === "ArrowUp") {
        event.preventDefault()
        this.#moveSelection(-1)
      } else if (event.key === "Escape") {
        this.#hideSuggestions()
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

    preventBlur(event) {
      event.preventDefault()
    }

    pickSuggestion(event) {
      const index = parseInt(event.currentTarget.dataset.index)
      this.#addLabel(this.#suggestions[index].name)
    }

    appendTag(label) {
      // Don't add duplicate tags.
      if (this.tagsTarget.querySelector(`[data-label-id="${label.ID}"]`)) return

      const span = document.createElement("span")
      span.className = "label-tag"
      span.dataset.labelId = label.ID
      span.innerHTML = `${this.#escapeHtml(label.Name)} <button class="label-remove" data-action="click->label-input#removeLabel" data-label-id="${label.ID}">&times;</button>`
      this.tagsTarget.appendChild(span)
    }

    #addLabel(name) {
      const url = this.urlValue
      if (!url) return

      fetch(url, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name })
      })
        .then(r => r.json())
        .then(label => {
          this.appendTag(label)
          this.inputTarget.value = ""
          this.#hideSuggestions()
        })
    }

    #fetchSuggestions(query) {
      fetch(`/api/labels?q=${encodeURIComponent(query)}`)
        .then(r => r.json())
        .then(labels => {
          this.#suggestions = labels
          this.#activeIndex = -1
          this.#renderSuggestions(query)
        })
    }

    #renderSuggestions(query) {
      if (this.#suggestions.length === 0) {
        this.#hideSuggestions()
        return
      }

      const html = this.#suggestions.map((label, i) => {
        const highlighted = this.#highlight(label.Name, query)
        const cls = i === this.#activeIndex ? "label-suggestion active" : "label-suggestion"
        return `<div class="${cls}" data-index="${i}" data-action="mousedown->label-input#preventBlur click->label-input#pickSuggestion">${highlighted}</div>`
      }).join("")

      this.suggestionsTarget.innerHTML = html
      this.suggestionsTarget.style.display = "block"
    }

    #hideSuggestions() {
      this.suggestionsTarget.style.display = "none"
      this.#suggestions = []
      this.#activeIndex = -1
    }

    #moveSelection(direction) {
      if (this.#suggestions.length === 0) return
      this.#activeIndex = Math.max(-1, Math.min(this.#suggestions.length - 1, this.#activeIndex + direction))

      const items = this.suggestionsTarget.querySelectorAll(".label-suggestion")
      items.forEach((el, i) => {
        el.classList.toggle("active", i === this.#activeIndex)
      })
    }

    #highlight(text, query) {
      const idx = text.toLowerCase().indexOf(query.toLowerCase())
      if (idx === -1) return this.#escapeHtml(text)
      const before = text.slice(0, idx)
      const match = text.slice(idx, idx + query.length)
      const after = text.slice(idx + query.length)
      return `${this.#escapeHtml(before)}<mark>${this.#escapeHtml(match)}</mark>${this.#escapeHtml(after)}`
    }

    #escapeHtml(str) {
      const div = document.createElement("div")
      div.textContent = str
      return div.innerHTML
    }
  }

  window.StimulusApp.register("label-input", LabelInputController)
})()
