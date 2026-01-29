import type { NextApiRequest, NextApiResponse } from 'next';
import { prisma } from '@/utils/prisma';
import { generatePermission } from '@/routes';

function parseBasicAuth(authHeader?: string) {
  if (!authHeader) return null;
  const [type, token] = authHeader.split(' ');
  if (type !== 'Basic' || !token) return null;
  try {
    const decoded = Buffer.from(token, 'base64').toString('utf8');
    const idx = decoded.indexOf(':');
    if (idx === -1) return null;
    const userName = decoded.slice(0, idx);
    return { userName };
  } catch {
    return null;
  }
}

export default async function handler(req: NextApiRequest, res: NextApiResponse) {
  if (req.method !== 'GET') {
    res.setHeader('Allow', ['GET']);
    return res.status(405).json({ msg: `Method ${req.method} Not Allowed` });
  }

  try {
    const { userName } = req.query || {};
    const auth = parseBasicAuth(req.headers.authorization);

    const lookupUserName =
      (typeof userName === 'string' && userName) || auth?.userName;

    if (!lookupUserName) {
      return res.status(400).json({ msg: 'userName is required' });
    }

    const user = await prisma.user.findUnique({ where: { userName: lookupUserName } });
    if (!user) {
      return res.status(404).json({ msg: 'User not found' });
    }

    const role = user.role === 'admin' ? 'admin' : 'user';

    return res.status(200).json({
      name: user.nickname || user.userName,
      email: user.email || '',
      avatar:
        'https://lf1-xgcdn-tos.pstatp.com/obj/vcloud/vadmin/start.8e0e4855ee346a46ccff8ff3e24db27b.png',
      accountId: user.id,
      registrationTime: user.createdAt.toISOString(),
      permissions: generatePermission(role),
      role,
      userName: user.userName,
    });
  } catch {
    return res.status(500).json({ msg: 'Failed to fetch user info' });
  }
}
