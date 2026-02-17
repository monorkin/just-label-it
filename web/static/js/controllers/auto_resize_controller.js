(() => {
  const { Controller } = Stimulus

  class AutoResizeController extends Controller {
    connect() {
      this.resize()
    }

    resize() {
      const el = this.element
      el.style.height = "auto"
      el.style.height = el.scrollHeight + "px"
    }
  }

  window.StimulusApp.register("auto-resize", AutoResizeController)
})()
