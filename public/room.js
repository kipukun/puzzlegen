function haveIntersection(r1, r2) {
    return !(
        r2.x > r1.x + r1.width ||
        r2.x + r2.width < r1.x ||
        r2.y > r1.y + r1.height ||
        r2.y + r2.height < r1.y
    );
}

(function (root) {
    'use strict'
    root.url = root.location.href + "/img/"
    fetch(root.location.href + "/info")
        .then((resp) => resp.json())
        .then(function (d) {
            root.nx = d["NX"];
            root.ny = d["NY"];
        })
        .catch(function (error) {
            console.log(error)
        })
    // const grid = document.getElementsByClassName("gallery");
    var wsURI = window.location.href.replace("http", "ws") + "/relay"
    var ws = new WebSocket(wsURI)
    ws.onmessage = function (e) {
        // canvas.loadFromJSON(e.data)
        console.log(e.data)
    }
    var s = new Konva.Stage({
        container: 'c',
        width: 500,
        height: 400
    })
    var l = new Konva.Layer()
    s.add(l)
    l.on('dragmove', function (e) {
        var target = e.target;
        var targetRect = e.target.getClientRect();
        l.children.each(function (group) {
            // do not check intersection with itself
            if (group === target) {
                return;
            }
            if (haveIntersection(group.getClientRect(), targetRect)) {
                var g = new Konva.Group({
                    draggable: true
                });
                group.setAttrs({
                    draggable: false
                })
                target.setAttrs({
                    draggable: false
                })
                g.add(group)
                g.add(target)
                l.add(g)
                console.log("yeah")
            } else {
                return
            }
            // do not need to call layer.draw() here
            // because it will be called by dragmove action
        });
    });
    Konva.Image.fromURL(root.url + '1/1', function (i) {
        i.setAttrs({
            draggable: true,
            x: 50,
            y: 100,
        })
        i.on('mouseover', function () {
            document.body.style.cursor = 'pointer';
        });
        i.on('mouseout', function () {
            document.body.style.cursor = 'default';
        });
        l.add(i)
        l.batchDraw();
    });
    Konva.Image.fromURL(root.url + '2/1', function (i) {
        i.setAttrs({
            draggable: true,
        })
        i.on('mouseover', function () {
            document.body.style.cursor = 'pointer';
        });
        i.on('mouseout', function () {
            document.body.style.cursor = 'default';
        });
        l.add(i);
        l.batchDraw();
    });
    Konva.Image.fromURL(root.url + '3/1', function (i) {
        i.setAttrs({
            draggable: true,
            x: 200,
            y: 200
        })
        i.on('mouseover', function () {
            document.body.style.cursor = 'pointer';
        });
        i.on('mouseout', function () {
            document.body.style.cursor = 'default';
        });
        l.add(i);
        l.batchDraw();
    });

})(this);