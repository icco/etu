export const TRACKING_ID = "WGVFU";

export const pageview = url => {
  window.fathom("trackPageview", {
    page_location: url,
  });
};

export const event = ({ action, category, label, value }) => {
  window.fathom("event", action, {
    event_category: category,
    event_label: label,
    value: value,
  });
};
