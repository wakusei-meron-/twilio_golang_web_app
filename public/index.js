// 必要な要素の取得
const setupButton = document.getElementById("js-setup-button");
const usernameElement = document.getElementById("js-username");
const phoneNumberElement = document.getElementById("js-phone-number-input");
const outgoingButton = document.getElementById("js-outgoing-button");
const incomingButton = document.getElementById("js-incoming-button");

// ステータスの定義
const Status = {
    Setup: 1,
    Outgoing: 2,
    Incoming: 3,
    IncomingAccept: 4,
}

// Twilioのセットアップ
let device
let outgoingCall
let incomingCall


// トークンの取得とTwilioクライアントの初期化
setupButton.onclick = async (e) => {
    const resp = await fetch("http://localhost:8686/token")
    const token = await resp.text()

    // twilioのSDKを初期化
    device = new Twilio.Device(token, {
        logLevel: 1,
        answerOnBridge: true,
        enableRingingState: true, edge: 'tokyo'
    })

    // twilioのイベントのコールバック時の挙動定義と登録
    addDeviceListeners(device);
    device.register()

    // 発行したトークンからログイン名の取得
    const payload = token.split(".")[1]
    const claims = JSON.parse(atob(payload))
    usernameElement.textContent = claims.grants.identity
}

// 発信・切電ボタン
outgoingButton.onclick = async (e) => {
    // 架電時のアクション
    if (outgoingCall) {
        outgoingCall.disconnect()
        outgoingCall = undefined
        updateUIAndState(Status.Setup)
        return
    }

    outgoingCall = await device.connect({params: {To: phoneNumberElement.value}})
    outgoingCall.on("accept", () => updateUIAndState(Status.Outgoing))
    outgoingCall.on("disconnect", () => {updateUIAndState(Status.Setup)})
}

// twilioのコールバックの定義
addDeviceListeners = (device) => {
    // 着信準備完了
    device.on('registered', () => updateUIAndState(Status.Setup))

    // 着信
    device.on("incoming", handleIncomingCall);
}

// UIと変数の更新
updateUIAndState = (status) => {
    switch (status) {
        case Status.Setup:
            outgoingCall = undefined
            incomingCall = undefined

            outgoingButton.disabled = false
            outgoingButton.textContent = "発信"
            incomingButton.disabled = true
            incomingButton.textContent = "電話に出る"
            break
        case Status.Outgoing:
            outgoingButton.textContent = "切電"
            break
        case Status.Incoming:
            incomingButton.disabled = false
            break
        case Status.IncomingAccept:
            incomingButton.textContent = "電話を切る"
            break
    }
}

// 着信時の処理
handleIncomingCall = (call) => {
    updateUIAndState(Status.Incoming)

    incomingButton.onclick = () => {
        if (incomingCall) {
            incomingCall.disconnect()
            return
        }
        call.accept()
        incomingCall = call
        updateUIAndState(Status.IncomingAccept)
    }

    call.on("cancel", () => updateUIAndState(Status.Setup)) // 電話に出ず発信元が電話を切った時
    call.on("disconnect", () => updateUIAndState(Status.Setup)) // 通話終了時
}


