import dynamic from "next/dynamic";

const P5Wrapper = dynamic(import("react-p5-wrapper"), {
  loading: () => "",
  ssr: false,
});

// spinning logo
function sketch(size) {
  return p => {
    let t = 0;
    let rand = [];

    p.setup = function() {
      p.createCanvas(size, size);

      p.noStroke();
      p.fill(51);
      rand = [
        p.random(-180, 180),
        p.random(-180, 180),
        p.random(-180, 180),
        p.random(-180, 180),
      ];
    };

    p.draw = function() {
      if (!(round(t) % 12)) {
        p.background(256, 60);
      }

      var k = p.width / 4;

      [[k * 1, k * 1], [k * 3, k * 1], [k * 1, k * 3], [k * 3, k * 3]].forEach(
        function(arr, i) {
          let x = arr[0];
          let y = arr[1];
          let r = rand[i];

          // each particle moves in a circle
          let myX = x + 0.15 * p.width * p.cos(2 * p.PI * t + r);
          let myY = y + 0.15 * p.width * p.sin(2 * p.PI * t + r);

          p.ellipse(myX, myY, 0.04 * p.width); // draw particle
        }
      );

      t = t + 0.01; // update time
    };
  };
}

function round(x) {
  return Number.parseFloat(x).toFixed(4);
}

const Link = params => {
  let size = 200;
  if (params.size) {
    size = params.size;
  }

  return (
    <div
      className={params.className}
      style={{ width: `${size}px`, height: `${size}px` }}
    >
      <P5Wrapper sketch={sketch(size)} />
    </div>
  );
};

export default Link;
