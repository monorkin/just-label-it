(() => {
  const { Controller } = Stimulus

  class NavigationController extends Controller {
    static values = { prevUrl: String, nextUrl: String }

    navigate(event) {
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
