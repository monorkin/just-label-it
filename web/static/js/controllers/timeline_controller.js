(() => {
  const { Controller } = Stimulus

  class TimelineController extends Controller {
    static targets = ["track", "playhead", "keyframe", "detail", "detailTime", "detailLabels", "detailDescription", "deleteBtn", "labelSection", "media"]
    static values = { fileId: Number }

    connect() {
      this._selectedId = null
      this._duration = 0
      this._animFrame = null
      this._scrubbing = false

      // If we have a media target, follow playback.
      if (this.hasMediaTarget) {
        this._onLoadedMetadata = () => {
          this._duration = this.mediaTarget.duration * 1000
          this._positionKeyframes()
        }
        this._onTimeUpdate = () => this._updatePlayhead()

        this.mediaTarget.addEventListener("loadedmetadata", this._onLoadedMetadata)
        this.mediaTarget.addEventListener("timeupdate", this._onTimeUpdate)

        // If metadata is already loaded.
        if (this.mediaTarget.duration) {
          this._duration = this.mediaTarget.duration * 1000
          this._positionKeyframes()
        }
      }

      // Bind scrub handlers for pointer drag on the track.
      this._onPointerMove = this._scrubMove.bind(this)
      this._onPointerUp = this._scrubEnd.bind(this)
    }

    disconnect() {
      if (this.hasMediaTarget) {
        this.mediaTarget.removeEventListener("loadedmetadata", this._onLoadedMetadata)
        this.mediaTarget.removeEventListener("timeupdate", this._onTimeUpdate)
      }
      document.removeEventListener("pointermove", this._onPointerMove)
      document.removeEventListener("pointerup", this._onPointerUp)
      cancelAnimationFrame(this._animFrame)
    }

    // --- Scrubbing (click / drag on timeline track) ---

    scrubStart(event) {
      // Ignore clicks on keyframe dots.
      if (event.target.classList.contains("timeline-keyframe")) return
      if (!this._duration) return

      this._scrubbing = true
      this._wasPlaying = !this.mediaTarget.paused
      this.mediaTarget.pause()

      this._seekToPointer(event)
      this.trackTarget.setPointerCapture(event.pointerId)
      document.addEventListener("pointermove", this._onPointerMove)
      document.addEventListener("pointerup", this._onPointerUp)
    }

    _scrubMove(event) {
      if (!this._scrubbing) return
      this._seekToPointer(event)
    }

    _scrubEnd(event) {
      if (!this._scrubbing) return
      this._scrubbing = false
      document.removeEventListener("pointermove", this._onPointerMove)
      document.removeEventListener("pointerup", this._onPointerUp)
      if (this._wasPlaying) this.mediaTarget.play()
    }

    _seekToPointer(event) {
      const rect = this.trackTarget.getBoundingClientRect()
      const x = event.clientX - rect.left
      const ratio = Math.max(0, Math.min(1, x / rect.width))
      this.mediaTarget.currentTime = (ratio * this._duration) / 1000
      this._updatePlayhead()
    }

    // --- Keyframe creation (separate button) ---

    addKeyframe() {
      if (!this._duration || !this.hasMediaTarget) return
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
          dot.dataset.action = "click->timeline#selectKeyframe"
          dot.title = kf.TimestampMs + "ms"
          this.trackTarget.appendChild(dot)
          this._positionKeyframes()
          this._selectKeyframe(kf.ID)
        })
    }

    selectKeyframe(event) {
      event.stopPropagation()
      const id = parseInt(event.currentTarget.dataset.keyframeId)
      this._selectKeyframe(id)

      // Seek media to keyframe time.
      if (this.hasMediaTarget) {
        const ms = parseInt(event.currentTarget.dataset.timestampMs)
        this.mediaTarget.currentTime = ms / 1000
      }
    }

    deleteKeyframe() {
      if (!this._selectedId) return

      const kfEl = this._findKeyframeEl(this._selectedId)
      if (kfEl && kfEl.dataset.pinned === "true") return

      fetch(`/keyframes/${this._selectedId}`, { method: "DELETE" }).then(r => {
        if (r.ok) {
          if (kfEl) kfEl.remove()
          this._selectedId = null
          if (this.hasDetailTarget) this.detailTarget.style.display = "none"
        }
      })
    }

    _selectKeyframe(id) {
      this._selectedId = id

      // Highlight the selected keyframe.
      this.keyframeTargets.forEach(el => {
        el.classList.toggle("selected", parseInt(el.dataset.keyframeId) === id)
      })

      const kfEl = this._findKeyframeEl(id)
      if (!kfEl) return

      const isPinned = kfEl.dataset.pinned === "true"
      const ms = parseInt(kfEl.dataset.timestampMs)

      // Show detail panel.
      if (this.hasDetailTarget) {
        this.detailTarget.style.display = ""
      }
      if (this.hasDetailTimeTarget) {
        this.detailTimeTarget.textContent = this._formatTime(ms)
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

      // Fetch keyframe details (labels + description).
      this._loadKeyframeDetail(id)
    }

    _loadKeyframeDetail(id) {
      // We don't have a dedicated endpoint, so we reconstruct from what we know.
      // For a full implementation, we'd add a GET /keyframes/{id} endpoint.
      // For now, we reset the UI and let the user type.
      if (this.hasDetailLabelsTarget) {
        this.detailLabelsTarget.innerHTML = ""
      }
      if (this.hasDetailDescriptionTarget) {
        this.detailDescriptionTarget.value = ""
        this.detailDescriptionTarget.style.height = "auto"
      }
    }

    _findKeyframeEl(id) {
      return this.keyframeTargets.find(el => parseInt(el.dataset.keyframeId) === id)
    }

    _positionKeyframes() {
      if (!this._duration) return
      this.keyframeTargets.forEach(el => {
        const ms = parseInt(el.dataset.timestampMs)
        const pct = (ms / this._duration) * 100
        el.style.left = pct + "%"
      })
    }

    _updatePlayhead() {
      if (!this.hasPlayheadTarget || !this.hasMediaTarget || !this._duration) return
      const currentMs = this.mediaTarget.currentTime * 1000
      const pct = (currentMs / this._duration) * 100
      this.playheadTarget.style.left = pct + "%"
    }

    _formatTime(ms) {
      const totalSec = Math.floor(ms / 1000)
      const min = Math.floor(totalSec / 60)
      const sec = totalSec % 60
      const msRemain = ms % 1000
      return `${min}:${String(sec).padStart(2, "0")}.${String(msRemain).padStart(3, "0")}`
    }
  }

  window.StimulusApp.register("timeline", TimelineController)
})()
