(() => {
  const application = Stimulus.Application.start()

  // Make application globally accessible for controller registration.
  window.StimulusApp = application
})()
