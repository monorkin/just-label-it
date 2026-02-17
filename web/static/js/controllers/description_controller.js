(() => {
  const { Controller } = Stimulus

  class DescriptionController extends Controller {
    static values = { url: String }

    connect() {
      this._timeout = null
    }

    disconnect() {
      clearTimeout(this._timeout)
    }

    save() {
      clearTimeout(this._timeout)
      this._timeout = setTimeout(() => {
        this._persist()
      }, 500)
    }

    _persist() {
      const url = this.urlValue
      if (!url) return

      fetch(url, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ description: this.element.value })
      })
    }
  }

  window.StimulusApp.register("description", DescriptionController)
})()
