let localStream;
let pcMap = new Map()
let socket;

const iceConfig = {
    iceServers: [{
        urls: 'stun:stun.l.google.com:19302'
    }, {
        urls: 'turn:8.210.83.164:3478', username: 'root', credential: 'root'
    }]
}

navigator.mediaDevices.getUserMedia({
    audio: {
        echoCancellationType: 'system',
        echoCancellation: true,
        noiseSuppression: true,
        autoGainControl: false,
        sampleRate: 24000,
        sampleSize: 16,
        channelCount: 2,
        volume: 0.5
    }, video: true,
}).then(function (stream) {
    let cid = getCid()
    let video = document.getElementById(cid)
    if (!video) {
        video = document.createElement("video")
        video.id = cid
        video.autoplay = true
        video.controls = true
        document.body.appendChild(video)
    }
    video.srcObject = stream
    localStream = stream
    console.log("local stream: ", stream)
}).catch(function (e) {
    console.error(e)
})

function join() {
    socket = new WebSocket(`wss://${window.location.host}/v1/webrtc/socket`)
    socket.onmessage = e => {
        console.log(e)

        let answerPc;
        let msg = JSON.parse(e.data)
        if (!msg) {
            return console.log('failed to parse msg')
        }

        let cid = getCid()
        if (msg.from === cid || msg.to !== cid) {
            return
        }

        switch (msg.type) {
            case "peers":
                if (pcMap.has(msg.cid)) {
                    document.getElementById(msg.cid).remove()
                }

                msg.data.peers.forEach(function (peer) {
                    const offerPc = new RTCPeerConnection(iceConfig);
                    const answerPc = new RTCPeerConnection(iceConfig);
                    pcMap.set(peer.cid, {
                        offer: offerPc, answer: answerPc
                    })

                    offerPc.onicecandidate = e => {
                        if (e.candidate && e.candidate.candidate !== "") {
                            socket.send(JSON.stringify({
                                "from": cid, "to": peer.cid, "answer": true, "type": "candidate", "data": e.candidate
                            }))
                        }
                    }
                    offerPc.oniceconnectionstatechange = () => {
                        console.log("peers offer oniceconnectionstatechange", offerPc.iceConnectionState)
                    }
                    offerPc.addTransceiver('video', {direction: 'sendonly'})
                    offerPc.addTransceiver('audio', {direction: 'sendonly'})
                    localStream.getTracks().forEach(function (track) {
                        console.log("peers offer add track: ", track)
                        offerPc.addTrack(track, localStream)
                    })

                    offerPc.createOffer().then(function (d) {
                        offerPc.setLocalDescription(d).then(function () {
                            socket.send(JSON.stringify({
                                "from": cid, "to": peer.cid, "answer": false, "type": "offer", "data": d
                            }))
                        })
                    })

                    answerPc.onicecandidate = e => {
                        if (e.candidate && e.candidate.candidate !== "") {
                            socket.send(JSON.stringify({
                                "from": cid, "to": peer.cid, "answer": false, "type": "candidate", "data": e.candidate
                            }))
                        }
                    }
                    answerPc.oniceconnectionstatechange = () => {
                        console.log("peers answer oniceconnectionstatechange", answerPc.iceConnectionState)
                        if (answerPc.iceConnectionState === "disconnected") {
                            const video = document.getElementById(peer.cid)
                            if (video.srcObject.active) {
                                return
                            }
                            video.remove()
                            console.log("peers answer video removed")
                        }
                    }
                    answerPc.ontrack = e => {
                        console.log("peers answer ontrack", e)
                        let video = document.getElementById(peer.cid)
                        if (!video) {
                            video = document.createElement("video")
                            video.id = peer.cid
                            video.autoplay = true
                            video.controls = true
                            document.body.appendChild(video)
                        }
                        video.srcObject = e.streams[0]
                        let pcs = pcMap.get(peer.cid)
                        pcs["streamId"] = e.streams[0].id
                        pcMap.set(peer.cid, pcs)
                    }
                    answerPc.addTransceiver('video', {direction: 'recvonly'})
                    answerPc.addTransceiver('audio', {direction: 'recvonly'})
                })
                break
            case "offer":
                if (msg.answer) {
                    answerPc = pcMap.get(msg.from).answer;
                    answerPc.setRemoteDescription(new RTCSessionDescription(msg.data)).then(function () {
                        answerPc.createAnswer().then(function (d) {
                            answerPc.setLocalDescription(d).then(function () {
                                socket.send(JSON.stringify({
                                    "from": cid, "to": msg.from, "type": "answer", "data": d
                                }))
                            })
                        })
                    }).catch(function (e) {
                        console.log(e)
                    })
                    break
                }

                answerPc = new RTCPeerConnection(iceConfig);
                let offerPc = new RTCPeerConnection(iceConfig)
                pcMap.set(msg.from, {
                    "answer": answerPc, "offer": offerPc
                })

                answerPc.onicecandidate = e => {
                    if (e.candidate && e.candidate.candidate !== "") {
                        socket.send(JSON.stringify({
                            "from": cid, "to": msg.from, "answer": false, "type": "candidate", "data": e.candidate
                        }))
                    }
                }
                answerPc.oniceconnectionstatechange = () => {
                    console.log("offer answer oniceconnectionstatechange", answerPc.iceConnectionState)
                    if (answerPc.iceConnectionState === "disconnected") {
                        const video = document.getElementById(msg.from)
                        if (video.srcObject.active) {
                            return
                        }
                        video.remove()
                        console.log("offer answer video removed")
                    }
                }
                answerPc.ontrack = e => {
                    console.log("offer answer ontrack", e)
                    let video = document.getElementById(msg.from)
                    if (!video) {
                        video = document.createElement("video")
                        video.id = msg.from
                        video.autoplay = true
                        video.controls = true
                        document.body.appendChild(video)
                    }
                    video.srcObject = e.streams[0]
                    let pcs = pcMap.get(msg.from)
                    pcs["streamId"] = e.streams[0].id
                    pcMap.set(msg.from, pcs)
                }
                answerPc.addTransceiver('video', {direction: 'recvonly'})
                answerPc.addTransceiver('audio', {direction: 'recvonly'})


                answerPc.setRemoteDescription(new RTCSessionDescription(msg.data)).then(function () {
                    answerPc.createAnswer().then(function (d) {
                        answerPc.setLocalDescription(d).then(function () {
                            socket.send(JSON.stringify({
                                "from": cid, "to": msg.from, "type": "answer", "data": d
                            }))
                        })
                    })
                }).catch(function (e) {
                    console.log(e)
                })

                offerPc.onicecandidate = e => {
                    if (e.candidate && e.candidate.candidate !== "") {
                        socket.send(JSON.stringify({
                            "from": cid, "to": msg.from, "answer": true, "type": "candidate", "data": e.candidate
                        }))
                    }
                }
                offerPc.oniceconnectionstatechange = () => {
                    console.log("offer offer oniceconnectionstatechange", offerPc.iceConnectionState)
                }
                offerPc.addTransceiver('video', {direction: 'sendonly'})
                offerPc.addTransceiver('audio', {direction: 'sendonly'})

                localStream.getTracks().forEach(function (track) {
                    console.log("offer offer add track: ", track)
                    offerPc.addTrack(track, localStream)
                })

                offerPc.createOffer().then(function (d) {
                    offerPc.setLocalDescription(d).then(function () {
                        socket.send(JSON.stringify({
                            "from": cid, "to": msg.from, "answer": true, "type": "offer", "data": d
                        }))
                    })
                })

                break
            case "answer":
                pcMap.get(msg.from).offer.setRemoteDescription(new RTCSessionDescription(msg.data)).then(function () {
                    console.log("remote answer stream: ", msg)
                }).catch(function (e) {
                    console.log(e)
                })
                break
            case "candidate":
                let pc = pcMap.get(msg.from).offer
                if (msg.answer) {
                    pc = pcMap.get(msg.from).answer
                }
                pc.addIceCandidate(new RTCIceCandidate(msg.data)).catch(function (e) {
                    console.log(e)
                })
                break
        }
    }

    socket.onopen = () => {
        let interval = setInterval(function () {
            if (!localStream) {
                return
            }
            let cid = getCid()
            socket.send(JSON.stringify({
                "from": cid, "type": "join", "data": {},
            }))
            window.clearInterval(interval)
        }, 1000)
    }
}

join()

function getCid() {
    let cid = localStorage.getItem("webrtc_cid")
    if (!cid) {
        cid = randomString(32)
        localStorage.setItem("webrtc_cid", cid)
    }
    return cid
}

function randomString(length) {
    const str = '0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ';
    let result = '';
    for (let i = length; i > 0; --i) result += str[Math.floor(Math.random() * str.length)];
    return result;
}