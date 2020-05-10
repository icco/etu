import Error from 'next/error'
import fetch from 'node-fetch'

export default function Page({ errorCode }) {
  const statusCode = errorCode ? errorCode : 404;
  return <Error statusCode={statusCode} />
}
