(() => {
  const { Controller } = Stimulus

  class TimelineController extends Controller {
    static targets = ["track", "playhead", "keyframe", "detail", "detailTime", "detailLabels", "detailDescription", "deleteBtn", "labelSection", "media"]
    static values = { fileId: Number }

    #selectedId = null
    #duration = 0
    #scrubbing = false
    #wasPlaying = false

    connect() {
      // If metadata is already loaded (cached media).
      if (this.hasMediaTarget && this.mediaTarget.duration) {
        this.#duration = this.mediaTarget.duration * 1000
        this.#positionKeyframes()
        this.#autoSelectFirst()
      }
    }

    // --- Media events (bound via data-action on <video>/<audio>) ---

    initializeTimeline() {
      this.#duration = this.mediaTarget.duration * 1000
      this.#positionKeyframes()
      this.#autoSelectFirst()
    }

    updatePlayhead() {
      if (!this.hasPlayheadTarget || !this.hasMediaTarget || !this.#duration) return
      const currentMs = this.mediaTarget.currentTime * 1000
      const pct = (currentMs / this.#duration) * 100
      this.playheadTarget.style.left = pct + "%"
    }

    // --- Scrubbing (click / drag on timeline track) ---

    startScrub(event) {
      // Ignore clicks on keyframe dots.
      if (event.target.classList.contains("timeline-keyframe")) return
      if (!this.#duration) return

      this.#scrubbing = true
      this.#wasPlaying = !this.mediaTarget.paused
      this.mediaTarget.pause()

      this.#seekToPointer(event)
      this.trackTarget.setPointerCapture(event.pointerId)
    }

    scrub(event) {
      if (!this.#scrubbing) return
      this.#seekToPointer(event)
    }

    endScrub() {
      if (!this.#scrubbing) return
      this.#scrubbing = false
      if (this.#wasPlaying) this.mediaTarget.play()
    }

    #seekToPointer(event) {
      const rect = this.trackTarget.getBoundingClientRect()
      const x = event.clientX - rect.left
      const ratio = Math.max(0, Math.min(1, x / rect.width))
      this.mediaTarget.currentTime = (ratio * this.#duration) / 1000
      this.updatePlayhead()
    }

    // --- Keyframe creation (separate button) ---

    addKeyframe() {
      if (!this.#duration || !this.hasMediaTarget) return
      const timestampMs = Math.round(this.mediaTarget.currentTime * 1000)

      fetch(`/files/${this.fileIdValue}/keyframes`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ timestamp_ms: timestampMs })
      })
        .then(r => r.json())
        .then(kf => {
          const dot = document.createElement("div")
          dot.className = "timeline-keyframe"
          dot.dataset.timelineTarget = "keyframe"
          dot.dataset.keyframeId = kf.ID
          dot.dataset.timestampMs = kf.TimestampMs
          dot.dataset.pinned = "false"
          dot.dataset.description = ""
          dot.dataset.labels = "[]"
          dot.dataset.action = "click->timeline#selectKeyframe"
          dot.title = kf.TimestampMs + "ms"
          this.trackTarget.appendChild(dot)
          this.#positionKeyframes()
          this.#selectKeyframe(kf.ID)
        })
    }

    selectKeyframe(event) {
      event.stopPropagation()
      const id = parseInt(event.currentTarget.dataset.keyframeId)
      this.#selectKeyframe(id)

      // Seek media to keyframe time.
      if (this.hasMediaTarget) {
        const ms = parseInt(event.currentTarget.dataset.timestampMs)
        this.mediaTarget.currentTime = ms / 1000
      }
    }

    deleteKeyframe() {
      if (!this.#selectedId) return

      const kfEl = this.#findKeyframeEl(this.#selectedId)
      if (kfEl && kfEl.dataset.pinned === "true") return

      fetch(`/keyframes/${this.#selectedId}`, { method: "DELETE" }).then(r => {
        if (r.ok) {
          if (kfEl) kfEl.remove()
          this.#selectedId = null
          if (this.hasDetailTarget) this.detailTarget.style.display = "none"
        }
      })
    }

    #selectKeyframe(id) {
      // Sync current keyframe state back to data attributes before switching.
      this.#syncCurrentKeyframe()

      this.#selectedId = id

      // Highlight the selected keyframe.
      this.keyframeTargets.forEach(el => {
        el.classList.toggle("selected", parseInt(el.dataset.keyframeId) === id)
      })

      const kfEl = this.#findKeyframeEl(id)
      if (!kfEl) return

      const isPinned = kfEl.dataset.pinned === "true"
      const ms = parseInt(kfEl.dataset.timestampMs)

      // Show detail panel.
      if (this.hasDetailTarget) {
        this.detailTarget.style.display = ""
      }
      if (this.hasDetailTimeTarget) {
        this.detailTimeTarget.textContent = this.#formatTime(ms)
      }
      if (this.hasDeleteBtnTarget) {
        this.deleteBtnTarget.style.display = isPinned ? "none" : ""
      }

      // Update the label section URL for this keyframe.
      if (this.hasLabelSectionTarget) {
        const labelController = this.application.getControllerForElementAndIdentifier(this.labelSectionTarget, "label-input")
        if (labelController) {
          labelController.urlValue = `/keyframes/${id}/labels`
        }
      }

      // Update description URL.
      if (this.hasDetailDescriptionTarget) {
        const descController = this.application.getControllerForElementAndIdentifier(this.detailDescriptionTarget, "description")
        if (descController) {
          descController.urlValue = `/keyframes/${id}/description`
        }
      }

      // Load keyframe details (labels + description).
      this.#loadKeyframeDetail(id)
    }

    #loadKeyframeDetail(id) {
      const kfEl = this.#findKeyframeEl(id)
      if (!kfEl) return

      // Load description from data attribute.
      if (this.hasDetailDescriptionTarget) {
        this.detailDescriptionTarget.value = kfEl.dataset.description || ""
        // Trigger auto-resize without firing input (which would trigger a save).
        const resizeCtrl = this.application.getControllerForElementAndIdentifier(this.detailDescriptionTarget, "auto-resize")
        if (resizeCtrl) resizeCtrl.resize()
      }

      // Load labels from data attribute.
      if (this.hasDetailLabelsTarget) {
        this.detailLabelsTarget.innerHTML = ""
        try {
          const labels = JSON.parse(kfEl.dataset.labels || "[]")
          const labelController = this.hasLabelSectionTarget
            ? this.application.getControllerForElementAndIdentifier(this.labelSectionTarget, "label-input")
            : null
          labels.forEach(label => {
            if (labelController) {
              labelController.appendTag(label)
            }
          })
        } catch (e) {
          // Ignore parse errors.
        }
      }
    }

    #autoSelectFirst() {
      if (this.#selectedId) return
      if (this.keyframeTargets.length > 0) {
        const firstId = parseInt(this.keyframeTargets[0].dataset.keyframeId)
        this.#selectKeyframe(firstId)
      }
    }

    #syncCurrentKeyframe() {
      if (!this.#selectedId) return
      const kfEl = this.#findKeyframeEl(this.#selectedId)
      if (!kfEl) return

      // Sync description.
      if (this.hasDetailDescriptionTarget) {
        kfEl.dataset.description = this.detailDescriptionTarget.value
      }

      // Sync labels from the tags container.
      if (this.hasDetailLabelsTarget) {
        const tags = this.detailLabelsTarget.querySelectorAll(".label-tag")
        const labels = Array.from(tags).map(tag => ({
          ID: parseInt(tag.dataset.labelId),
          Name: tag.childNodes[0].textContent.trim()
        }))
        kfEl.dataset.labels = JSON.stringify(labels)
      }
    }

    #findKeyframeEl(id) {
      return this.keyframeTargets.find(el => parseInt(el.dataset.keyframeId) === id)
    }

    #positionKeyframes() {
      if (!this.#duration) return
      this.keyframeTargets.forEach(el => {
        const ms = parseInt(el.dataset.timestampMs)
        const pct = (ms / this.#duration) * 100
        el.style.left = pct + "%"
      })
    }

    #formatTime(ms) {
      const totalSec = Math.floor(ms / 1000)
      const min = Math.floor(totalSec / 60)
      const sec = totalSec % 60
      const msRemain = ms % 1000
      return `${min}:${String(sec).padStart(2, "0")}.${String(msRemain).padStart(3, "0")}`
    }
  }

  window.StimulusApp.register("timeline", TimelineController)
})()
