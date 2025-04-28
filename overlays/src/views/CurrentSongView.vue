<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

interface TrackResponse {
    event: string;
    track: string;
    artists: string[];
    album: string;
    time: Date;
    image: {
      url: string;
    };
}

const track = ref<TrackResponse>({
    event: '',
    track: '',
    artists: [],
    album: '',
    time: new Date(Date.now()),
    image: {
      url: ''
    }
});

const titleRef = ref<HTMLElement | null>(null);
const albumNameRef = ref<HTMLElement | null>(null);
const artistsRef = ref<HTMLElement | null>(null);
let socket: WebSocket | null = null;

onMounted(() => {

  socket = new WebSocket('http://localhost:8080/api/ws/spotify-current-playing');
  socket.onmessage = (event) => {
    try {
      const data: TrackResponse = JSON.parse(event.data);

      if(data.track != null) {
        track.value = {
            ...data,
            time: new Date(data.time)
        };
      
        if (titleRef.value) animateElement(titleRef.value);
        if (albumNameRef.value) animateElement(albumNameRef.value);
        if (artistsRef.value) animateElement(artistsRef.value);
      }

    } catch (error) {
        console.error('Error parsing WebSocket message:', error);
    }
  };
})

onUnmounted(() => {

  if (socket) {
    socket.close();
    socket = null;
  }

  if (titleRef.value) {
    titleRef.value.classList.remove('scrolling')
  }
  if (albumNameRef.value) {
    albumNameRef.value.classList.remove('scrolling')
  }
  if (artistsRef.value) {
    artistsRef.value.classList.remove('scrolling')
  }
})

function animateElement(element: HTMLElement){
    const isOverflowing = element.scrollWidth > element.clientWidth
    
    if (!isOverflowing) {
        return
    }

    const scrollSpeed = 50
    const duration = Math.max(10, element.scrollWidth / scrollSpeed)
    element.style.setProperty('--animation-duration', `${duration}s`)
    element.classList.add('scrolling')
}

</script>

<template>
  <div class="obs-overlay transparent">
    <div class="content">
      <div class="data preview">
        <img
          :src="track.image.url"
          alt="Album cover"
        />
      </div>
      <div class="data song">
        <h1 ref="titleRef">{{ track.track }}</h1>
        <h2 ref="albumNameRef">{{ track.album }}</h2>
        <p ref="artistsRef">{{ track.artists.join(', ') }}</p>
      </div>
    </div>
  </div>
</template>

<style>
.obs-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
  color: white;
  font-family: 'Arial', 'Helvetica', sans-serif;
  overflow: hidden;
}

.transparent {
  background-color: transparent !important;
}

.content {
  width: min(50%, 600px);
  height: 20vh;
  background-color: rgba(0, 0, 0, 0.7);
  border-radius: 5px;
  display: flex;
  justify-content: center;
  align-items: center;
  flex-direction: row;
  padding: 10px;
  border: 2px solid rgba(255, 255, 255, 0.2);
}

.data {
  display: flex;
  justify-content: center;
  width: 50%;
  height: 100%;
  overflow: hidden;
  padding: 10px;
}

.preview img {
  max-width: 100%;
  max-height: 100%;
  width: auto;
  height: auto;
  object-fit: contain;
  display: block;
  border-radius: 3px;
}

.song {
  justify-content: center;
  flex-direction: column;
  gap: 2px;
  overflow: hidden;
  flex: 1;
}

.song h1 {
  margin: 0 0 2px 0;
  font-size: 1.2em;
  white-space: nowrap;
  color: white;
}

.song h2 {
  margin: 0 0 2px 0;
  font-size: 1em;
  color: #ccc;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.song p {
  margin: 0;
  font-size: 0.9em;
  color: #aaa;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.scrolling {
  padding-right: 20px;
  animation: scrollText var(--animation-duration, 10s) linear infinite;
}

@keyframes scrollText {
  0% {
    transform: translateX(0);
  }
  100% {
    transform: translateX(calc(-100% + 20px));
  }
}

@media (max-width: 600px) {
  .content {
    width: 90%;
  }
  .song h1 {
    font-size: 1em;
  }
  .song h2 {
    font-size: 0.9em;
  }
  .song p {
    font-size: 0.8em;
  }
}
</style>