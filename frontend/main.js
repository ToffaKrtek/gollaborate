import { EditorState,  } from "@codemirror/state";
import { EditorView } from "@codemirror/view";
import { javascript } from "@codemirror/lang-javascript";
import * as Y from "yjs";

class WebSocketProvider {
    constructor(url, doc, onSync) {
        this.ws = new WebSocket(url);
        this.ws.binaryType = "arraybuffer";
        this.ws.onopen = () => onSync?.(true);
        this.ws.onmessage = (event) => {
            const update = new Uint8Array(event.data);
            Y.applyUpdate(doc, update);
        };
        this.ws.onclose = () => onSync?.(false);
        doc.on("update", (update) => {
            if (this.ws.readyState === WebSocket.OPEN) {
                this.ws.send(update);
            }
        });
    }
    disconnect() { this.ws.close(); }
}

const urlParams = new URLSearchParams(window.location.search);
const docName = urlParams.get('doc') || 'untitled';
const wsUrl = `ws://${window.location.host}/ws?doc=${docName}`;

const ydoc = new Y.Doc();
const provider = new WebSocketProvider(wsUrl, ydoc, (synced) => console.log(synced ? "synced" : "offline"));
const ytext = ydoc.getText("codemirror");

const state = EditorState.create({
    doc: ytext.toString(),
    extensions: [
        javascript(),
        EditorView.updateListener.of((update) => {
            if (update.docChanged) {
                const newVal = update.state.doc.toString();
                if (newVal !== ytext.toString()) {
                    ytext.delete(0, ytext.length);
                    ytext.insert(0, newVal);
                }
            }
        }),
    ],
});

const view = new EditorView({ state, parent: document.getElementById("editor") });

ytext.observe(() => {
    const current = view.state.doc.toString();
    const yval = ytext.toString();
    if (current !== yval) {
        view.dispatch({ changes: { from: 0, to: current.length, insert: yval } });
    }
});
