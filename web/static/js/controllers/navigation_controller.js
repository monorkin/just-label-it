(() => {
  const { Controller } = Stimulus

  class NavigationController extends Controller {
    static values = { prevUrl: String, nextUrl: String }

    connect() {
      this._onKeydown = this._handleKeydown.bind(this)
      document.addEventListener("keydown", this._onKeydown)
    }

    disconnect() {
      document.removeEventListener("keydown", this._onKeydown)
    }

    _handleKeydown(event) {
      // Don't navigate when typing in an input or textarea.
      const tag = event.target.tagName
      if (tag === "INPUT" || tag === "TEXTAREA") return

      if (event.key === "ArrowLeft") {
        event.preventDefault()
        window.location.href = this.prevUrlValue
      } else if (event.key === "ArrowRight") {
        event.preventDefault()
        window.location.href = this.nextUrlValue
      }
    }
  }

  window.StimulusApp.register("navigation", NavigationController)
})()
