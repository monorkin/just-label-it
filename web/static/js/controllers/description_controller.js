(() => {
  const { Controller } = Stimulus

  class DescriptionController extends Controller {
    static values = { url: String }

    #timeout = null

    disconnect() {
      clearTimeout(this.#timeout)
    }

    save() {
      clearTimeout(this.#timeout)
      this.#timeout = setTimeout(() => {
        this.#persist()
      }, 500)
    }

    #persist() {
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
